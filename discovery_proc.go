// Discovers namespaces from the /proc filesystem. This discovery runs only
// once in the current PID namespace. The rationale is two-fold: first, we
// already see all processes, even those in child PID namespaces (because PID
// namespaces are not only hierarchical, but also nested). And second, the
// Linux kernel blocks entering parent PID (and user) namespaces; and in
// consequence of this rule, the kernel also blocks entering sibling PID
// namespaces.
//
// See also: http://man7.org/linux/man-pages/man7/pid_namespaces.7.html,
// http://man7.org/linux/man-pages/man2/setns.2.html (clearly spelling out the
// rules), as well as
// http://man7.org/linux/man-pages/man7/user_namespaces.7.html (yes, user
// namespaces are also important in the whole game).

// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux

package lxkns

import (
	"fmt"
	"os"

	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

// discoverFromProc discovers Linux kernel namespaces from the process table,
// using the namespace links inside the proc filesystem: "/proc/[PID]/ns/...".
// It does not check any other places, as these are covered by separate
// discovery functions.
func discoverFromProc(nstype species.NamespaceType, _ string, result *DiscoveryResult) {
	if !result.Options.ScanProcs {
		log.Infof("skipping discovery of %s namespaces used by processes", nstype.Name())
		return
	}
	log.Debugf("discovering %s namespaces used by processes...", nstype.Name())
	total := 0
	nstypename := nstype.Name()
	nstypeidx := model.TypeIndex(nstype)
	nsmap := result.Namespaces[nstypeidx]
	// For all processes (but not tasks/threads) listed in /proc try to gather
	// the namespaces of a given type they use.
	for pid, proc := range result.Processes {
		// Discover the namespace instance of the specified type which this
		// particular process has joined. Please note that namespace
		// references for processes appear as symbolic(!) links in the /proc
		// filesystem, but in fact are behaving like hard links. Nevertheless,
		// we have to follow them like symbolic links in order to find the
		// identifier in form of the inode # of the referenced namespace.
		nsref := fmt.Sprintf("/proc/%d/ns/%s", pid, nstypename)
		// Avoid using high-level golang i/o calls, as these like to hand over
		// to yet another goroutine, something which really doesn't help us
		// here. Please note that we need the open fd further below in case we
		// need to discover ownership.
		f, err := os.Open(nsref)
		if err != nil {
			continue
		}
		// Why not using a (typed) NamespacePath here? Because we want to carry
		// out multiple query operations and avoid repeated opening and closing
		// for each single query.
		nsf, _ := ops.NewTypedNamespaceFile(f, nstype)
		nsid, err := nsf.ID()
		if err != nil {
			nsf.Close() // ...don't leak!
			continue
		}
		ns, ok := nsmap[nsid]
		if !ok {
			// Only add a namespace we haven't yet seen. And yes, we don't
			// give a reference here, as we want to use a reference from a
			// leader process, and not of some child process deep down the
			// hierarchy, which might not even live for long (as sad as this
			// might be).
			ns = namespaces.New(nstype, nsid, "")
			nsmap[nsid] = ns
			log.Debugf("found namespace %s", ns.(model.NamespaceStringer).TypeIDString())
			total++
		}
		// To speed up finding the process leaders in a specific namespace, we
		// remember this namespace as joined by the process we're just looking
		// at. Additionally, other applications also benefit from quickly
		// navigating from a process to its joined namespace proxy objects.
		proc.Namespaces[nstypeidx] = ns
		// Let's also get the owning user namespace id, while we still have a
		// suitable fd open. For user namespaces, we skip this step, as this
		// is the same as the parent relationship. Additionally, it makes
		// things too awkward in the model, because then we would need to
		// treat ownership differently for non-user namespaces versus user
		// namespaces all the time. Thus, sorry, no user namespaces here.
		if result.Options.DiscoverOwnership && nstype != species.CLONE_NEWUSER {
			ns.(namespaces.NamespaceConfigurer).DetectOwner(nsf)
		}
		// Don't leak... And no, defer won't help us here.
		nsf.Close()
	}
	// Now that we know which namespaces are existing with processes joined to
	// them, let's find out the leader processes in these namespaces...
	for pid, proc := range result.Processes {
		// In case we got no access to this process, we must skip it. And we
		// must remove it from our process table, so others won't try to use
		// them. This will not remove the process from the process tree, rest
		// assured.
		if proc.Namespaces[nstypeidx] == nil {
			// Time namespaces are new since kernel 5.6, so many
			// deployments won't have a kernel which supports them. Don't
			// prune then, as we would end up with an empty process table :(
			if nstypeidx != model.TimeNS {
				delete(result.Processes, pid)
			}
			continue
		}
		// Find leader from this position in the process tree: a "leader" is
		// the topmost process in the process tree which is still joined to
		// the same namespace as the namespace of the process from which we
		// started our quest.
		p := proc
		parentp := p.Parent
		for parentp != nil && parentp.Namespaces[nstypeidx] == p.Namespaces[nstypeidx] {
			p = parentp
			parentp = p.Parent
		}
		p.Namespaces[nstypeidx].(namespaces.NamespaceConfigurer).AddLeader(p)
	}
	// Try to set namespace references which we hope to be as long-lived as
	// possible; so we prefer the most senior leader process: the ealdorman.
	for _, ns := range nsmap {
		if senior := ns.Ealdorman(); senior != nil {
			ns.(namespaces.NamespaceConfigurer).SetRef(
				fmt.Sprintf("/proc/%d/ns/%s", senior.PID, nstypename))
		}
	}

	log.Infof("found %d %s namespaces joined by processes", total, nstype.Name())
}
