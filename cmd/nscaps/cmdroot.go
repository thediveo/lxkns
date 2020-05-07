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
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
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

// nsnode represents a namespace node, and either a user namespace node or a
// non-user namespace node. A user namespace node can have child nodes. A
// non-user namespace node cannot have children. A namespace node can act as a
// target namespace.
type nsnode struct {
	ns       lxkns.Namespace
	istarget bool
	children []node
}

func (n *nsnode) Children() []node { return n.children }

// processnode represents the "reference" process whose capabilities are to be
// evaluated in a target namespace. A processnode always terminates a branch.
type processnode struct {
	proc *lxkns.Process
}

func (p *processnode) Children() []node { return []node{} }

// processbranch returns the branch from the initial user namespace down to the
// user namespace containing the specified process. So, the process branch
// completely consists of namespace nodes with a final process node.
func processbranch(proc *lxkns.Process) (n node) {
	// Branch always ends in a user namespace node with a process node as its
	// sole child.
	userns := proc.Namespaces[lxkns.UserNS].(lxkns.Ownership)
	n = &nsnode{
		ns: userns.(lxkns.Namespace),
		children: []node{
			&processnode{
				proc: proc,
			},
		},
	}
	// Now climb up the user namespace hierarchy, completing the branch
	// "upwards" towards the root. Each parent user namespace has a sole child,
	// its child user namespace.
	for userns.(lxkns.Hierarchy).Parent() != nil {
		userns = userns.(lxkns.Hierarchy).Parent().(lxkns.Ownership)
		n = &nsnode{
			ns: userns.(lxkns.Namespace),
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
func targetbranch(tns lxkns.Namespace) (n node) {
	var userns lxkns.Ownership
	if tns.Type() == species.CLONE_NEWUSER {
		// Please note that the lxkns namespace model on purpose does not set
		// the owner relationship on user namespaces: that's the parent
		// relationship instead.
		userns = tns.(lxkns.Ownership)
		n = &nsnode{
			ns:       userns.(lxkns.Namespace),
			istarget: true,
		}
	} else {
		// Non-user namespaces have their owning user namespace relationship set
		// in the lxkns information model.
		userns = tns.Owner()
		n = &nsnode{
			ns: userns.(lxkns.Namespace),
			children: []node{
				&nsnode{
					ns:       tns,
					istarget: true,
				},
			},
		}
	}
	// Now climb up the user namespace hierarchy, completing the branch
	// "upwards" towards the root. Each parent user namespace has a sole child,
	// its child user namespace.
	for userns.(lxkns.Hierarchy).Parent() != nil {
		userns = userns.(lxkns.Hierarchy).Parent().(lxkns.Ownership)
		n = &nsnode{
			ns: userns.(lxkns.Namespace),
			children: []node{
				n,
			},
		}
	}
	return
}

// fork combines the process branch with the target namespace branch to the
// extend that these branches share commong user namespaces, or even a target
// user namespace.
func fork(pbr node, tbr node) (root node) {
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
			ppbr.children = append(ppbr.children, tbr)
			break
		}
		if pnsnode.ns != tbr.(*nsnode).ns {
			// The target branch forks off here; so add in the forking target
			// branch at our parent, and then let's call it a day ;)
			ppbr.children = append(ppbr.children, tbr)
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

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "nscaps [flags] NAMESPACE",
		Short:   "nscaps shows the capabilities of a process in a particular namespace",
		Version: lxkns.SemVersion,
		Args:    cobra.ExactArgs(1),
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.BeforeCommand()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nsid, nst := species.IDwithType(args[0])
			if nst == species.NaNS {
				return fmt.Errorf("not a valid namespace: %q", args[0])
			}
			fpid, _ := cmd.PersistentFlags().GetUint32("pid")
			pid := lxkns.PIDType(fpid)
			// Has a PID namespace different than our current one been
			// specified, in which the PID is valid?
			pidnsid := species.NoneID
			if nst, _ := cmd.PersistentFlags().GetString("ns"); nst != "" {
				id, err := strconv.ParseUint(nst, 10, 64)
				if err == nil {
					pidnsid, _ = species.IDwithType(fmt.Sprintf("pid:[%d]", id))
				} else {
					var t species.NamespaceType
					pidnsid, t = species.IDwithType(nst)
					if t != species.CLONE_NEWPID {
						return fmt.Errorf("not a valid PID namespace ID: %q", nst)
					}
				}
			}
			// Without a specific PID given, use our own PID as a substitution.
			if pid == 0 {
				if pidnsid != species.NoneID {
					return fmt.Errorf("--ns requires --pid=PID")
				}
				pid = lxkns.PIDType(os.Getpid())
			}
			// Run a full namespace discovery and also get the PID translation map.
			allns := lxkns.Discover(lxkns.FullDiscovery)
			pidmap := lxkns.NewPIDMap(allns)
			rootpidns := allns.Processes[lxkns.PIDType(os.Getpid())].Namespaces[lxkns.PIDNS]
			// If necessary, translate the PID from its own PID namespace into the
			// initial/this program's PID namespace.
			if pidnsid != species.NoneID {
				pidns := allns.Namespaces[lxkns.PIDNS][pidnsid]
				if pidns == nil {
					return fmt.Errorf("unknown PID namespace pid:[%d]", pidnsid.Ino)
				}
				rootpid := pidmap.Translate(lxkns.PIDType(pid), pidns, rootpidns)
				if rootpid == 0 {
					return fmt.Errorf("unknown process PID %d in pid:[%d]",
						pid, pidnsid.Ino)
				}
				pid = rootpid
			}
			//
			proc, ok := allns.Processes[pid]
			if !ok {
				return fmt.Errorf("unknown process PID %d", pid)
			}
			procbr := processbranch(proc)
			tns, ok := allns.Namespaces[lxkns.TypeIndex(nst)][nsid]
			if !ok {
				return fmt.Errorf("unknown namespace %s", args[0])
			}
			tbr := targetbranch(tns)
			root := fork(procbr, tbr)
			fmt.Printf("%+v", root)
			return nil
		},
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().Uint32P("pid", "p", 0,
		"PID of process for which to calculate capabilities")
	rootCmd.PersistentFlags().StringP("ns", "n", "",
		"PID namespace of PID, if not the initial PID namespace;\n"+
			"either an unsigned int64 value, such as \"4026531836\", or a\n"+
			"PID namespace textual representation like \"pid:[4026531836]\"")
	cli.AddFlags(rootCmd)
	return
}
