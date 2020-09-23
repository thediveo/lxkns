// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// nodes are namespaces which can have 0, 1, or 2 child nodes. Following the
// Linux kernel namespace model and especially the namespace capabilities model,
// in this particular situation, only user namespaces can have child nodes.
// Children can be either user namespaces, non-user namespaces (acting as
// "targets"), and finally process nodes. There can be at most one namespace
// child and one process child. Processes don't have child nodes in our
// particular model here.
type node interface {
	Children() []node
}

type targetcaps int

const (
	incapable targetcaps = iota
	effcaps
	allcaps
)

// nsnode represents a namespace node, and either a user namespace node or a
// non-user namespace node. A user namespace node can have child nodes. A
// non-user namespace node cannot have children. A namespace node can act as a
// target namespace.
type nsnode struct {
	ns         model.Namespace
	istarget   bool
	targetcaps targetcaps
	children   []node
}

func (n nsnode) Children() []node { return n.children }

// processnode represents the "reference" process whose capabilities are to be
// evaluated in a target namespace. A processnode always terminates a branch.
type processnode struct {
	proc *model.Process
	euid int
	caps []string
}

func (p processnode) Children() []node { return []node{} }

// processbranch returns the branch from the initial user namespace down to the
// user namespace containing the specified process. So, the process branch
// completely consists of namespace nodes with a final process node.
func processbranch(proc *model.Process, euid int) (n node, err error) {
	// Branch always ends in a user namespace node with a process node as its
	// sole child.
	userns, ok := proc.Namespaces[model.UserNS].(model.Ownership)
	if !ok { // actually, this means that the user namespace is really nil
		return nil, fmt.Errorf(
			"cannot access namespace information of process PID %d", proc.PID)
	}
	if euid < 0 {
		return nil, fmt.Errorf("cannot query effective UID of process PID %d",
			proc.PID)
	}
	caps := ProcessCapabilities(proc.PID)
	n = &nsnode{
		ns: userns.(model.Namespace),
		children: []node{
			&processnode{
				proc: proc,
				euid: euid,
				caps: caps,
			},
		},
	}
	// Now climb up the user namespace hierarchy, completing the branch
	// "upwards" towards the root. Each parent user namespace has a sole child,
	// its child user namespace.
	for userns.(model.Hierarchy).Parent() != nil {
		userns = userns.(model.Hierarchy).Parent().(model.Ownership)
		n = &nsnode{
			ns: userns.(model.Namespace),
			children: []node{
				n,
			},
		}
	}
	return
}

// targetbranch returns the branch from the initial user namespace down to the
// target namespace. Please note that for a user namespace target the branch
// ends in a that type of namespace, with istarget set. Otherwise, the branch
// ends in a non-user namespace node, again with istarget set. So, a target
// branch always consists only of namespace nodes, with the final one having its
// istarget flag set. All nodes, except maybe for the last, are user namespaces.
func targetbranch(tns model.Namespace, tcaps targetcaps) (n node) {
	var userns model.Ownership
	if tns.Type() == species.CLONE_NEWUSER {
		// Please note that the lxkns namespace model on purpose does not set
		// the owner relationship on user namespaces: that's the parent
		// relationship instead.
		userns = tns.(model.Ownership)
		n = &nsnode{
			ns:         userns.(model.Namespace),
			istarget:   true,
			targetcaps: tcaps,
		}
	} else {
		// Non-user namespaces have their owning user namespace relationship set
		// in the lxkns information model.
		userns = tns.Owner()
		n = &nsnode{
			ns: userns.(model.Namespace),
			children: []node{
				&nsnode{
					ns:         tns,
					istarget:   true,
					targetcaps: tcaps,
				},
			},
		}
	}
	// Now climb up the user namespace hierarchy, completing the branch
	// "upwards" towards the root. Each parent user namespace has a sole child,
	// its child user namespace.
	for userns.(model.Hierarchy).Parent() != nil {
		userns = userns.(model.Hierarchy).Parent().(model.Ownership)
		n = &nsnode{
			ns: userns.(model.Namespace),
			children: []node{
				n,
			},
		}
	}
	return
}

// combine the process branch with the target namespace branch to the extend
// that these branches share commong user namespaces, or even a target user
// namespace.
func combine(pbr node, tbr node) (root node) {
	// TODO: process and target branching not sharing a common root?!
	root = pbr
	// If you find a fork in the road ... take it! Please note that we here now
	// can rely on the fact that both branches always start with a user
	// namespace, which should be the (true or fake) initial namespace.
	ppbr := (*nsnode)(nil) // no parent process branch node yet.
	for {
		pnsnode, ok := pbr.(*nsnode)
		if !ok {
			// The process branch forks off here, as we've stumbled onto the
			// final process node in the process branch. Thus, we need to add
			// the target branch to our common parent user namespace node.
			//
			//      U
			//      |\
			// pbr->P ?
			ppbr.children = append(ppbr.children, tbr)
			// Mark the parent user namespace, as well as all as descendent user
			// name spaces down to the target namespace according to what extend
			// the process will have capabilities in them.
			ppbr.targetcaps = effcaps
			for tbr != nil {
				if tbr.(*nsnode).istarget {
					break
				}
				tbr.(*nsnode).targetcaps = allcaps
				tbr = tbr.Children()[0]
			}
			break
		}
		if pnsnode.ns != tbr.(*nsnode).ns {
			// The target branch forks off here; so add in the forking target
			// branch at our parent, and then let's call it a day ;)
			ppbr.children = append(ppbr.children, tbr)
			// Mark only the user namespace node containing the process node as
			// offering effective capabilities. Everything else is off limits.
			for {
				if _, ok := pbr.(*processnode); ok {
					ppbr.targetcaps = effcaps
					break
				}
				ppbr = pbr.(*nsnode)
				pbr = pbr.Children()[0]

			}
			break
		}
		// Both branches still share the same user namespace node. But make sure
		// to take over the istarget flag from the target branch, as this might
		// be the final node in the target branch.
		pnsnode.istarget = tbr.(*nsnode).istarget
		tbrch := tbr.Children()
		if len(tbrch) == 0 {
			// The target branch ends here, so we're done.
			break
		}
		tbr = tbrch[0] // remember: at most one child node.
		// At this point, we know that the current process branch node is a user
		// namespace node, and thus must have still one child node: either another
		// user namespace node, or a process node. So we can blindly take
		// whatever child we get, sure that there actually is a child.
		ppbr = pbr.(*nsnode)
		pbr = pbr.Children()[0]
	}
	return
}

// caps decides based on the specific process and target namespace if the
// process has no capabilities, its effective capabilities, or full capabilities
// (subject to Linux security modules).
func caps(proc *model.Process, tns model.Namespace) (tcaps targetcaps, euid int, err error) {
	// Get the user namespace the process is currently joined to. Since we're
	// later to climb the ladder, we're more interested in hierarchy than
	// ownership, just as people often behave. Another case where the things we
	// create very much look the same as we do.
	procuserns, ok := proc.Namespaces[model.UserNS].(model.Hierarchy)
	if !ok {
		return incapable, -1, fmt.Errorf(
			"cannot access namespace information of process PID %d", proc.PID)
	}
	// Get the user namespace owning the target namespace ... unless the target
	// namespace is a user namespace: then we use the target user namespace itself.
	var targetuserns model.Hierarchy
	if tns.Type() == species.CLONE_NEWUSER {
		targetuserns = tns.(model.Hierarchy)
	} else {
		targetuserns, ok = tns.Owner().(model.Hierarchy)
		if !ok {
			return incapable, -1, fmt.Errorf(
				"cannot access owning user namespace information of target namespace %s",
				tns.(model.NamespaceStringer).TypeIDString())
		}
	}
	euid = processEuid(proc)
	// If the process happens to be in the target (owning) user namespace, then
	// it has its effective caps. Now, that was to be expected, as otherwise how
	// could the process have been there anyway?
	//
	// See also rule #1 of
	// http://man7.org/linux/man-pages/man7/user_namespaces.7.html: "A process
	// has a capability inside a user namespace if it is a member of that
	// namespace and it has the capability in its effective capability set"
	if procuserns == targetuserns {
		return effcaps, euid, nil
	}
	// The target user namespace must be below the process' user namespace,
	// albeit this isn't sufficient by itself. However, this covers rule #2
	// http://man7.org/linux/man-pages/man7/user_namespaces.7.html: "If a
	// process has a capability in a user namespace, then it has that capability
	// in all child (and further removed descendant) namespaces as well." So we
	// now climb up the user namespace hierarchy above the target's user
	// namespace, checking if we find the process' user namespace.
	userns := targetuserns
	childuserns := model.Hierarchy(nil)
	for {
		if userns == procuserns {
			break
		}
		childuserns = userns
		userns = userns.Parent()
		if userns == nil {
			// We've reached the top of the hierarchy, yet didn't find our
			// process' user namespace, so we're toast. No capabilities, totally
			// incapable. Oh, never mind.
			return incapable, euid, nil
		}
	}
	// Rule #3 basically sez: if the owner of the user namespace just
	// immediately below the process' user namespace on the path to the target's
	// user namespace has the same owner UID as the process' effective UID, then
	// we're in like Flynn ... with apologies to Dave. See
	// http://man7.org/linux/man-pages/man7/user_namespaces.7.html: "A process
	// that resides in the parent of the user namespace and whose effective user
	// ID matches the owner of the namespace has all capabilities in the
	// namespace. By virtue of the previous rule, this means that the process
	// has all capabilities in all further removed descendant user namespaces as
	// well."
	if euid == childuserns.(model.Ownership).UID() {
		return allcaps, euid, nil
	}
	// Rule #2, but not rule #3: target's user namespace below process, but
	// different owners: close, but no royal furcup, so effective caps only
	// (which might be sufficient anyway, if we're lucky).
	return effcaps, euid, nil
}

// processEuid returns the effective UID of the specified process (as opposed to
// os.Geteuid which gets the effective UID of the current process). Returns -1
// in case the effective UID cannot be determined.
func processEuid(proc *model.Process) int {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", proc.PID))
	if err != nil {
		return -1
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// Scan through the procfess status information until we arrive at the
	// "Uid:" field.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:\t") {
			uidtxts := strings.Split(line[5:], "\t")
			euid, err := strconv.Atoi(uidtxts[1])
			if err != nil {
				return -1
			}
			return euid
		}
	}
	panic("/proc filesystem broken: no Uid element in status.")
}
