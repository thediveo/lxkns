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

//go:build linux

package discover

import (
	"fmt"
	"os"

	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/plural"
	"github.com/thediveo/lxkns/species"
)

// discoverFromProc discovers Linux kernel namespaces from the process table
// (including tasks when requested), using the namespace links inside the proc
// filesystem: "/proc/[PID]/ns/...". It does not check any other places, as
// these are covered by separate discovery functions.
func discoverFromProc(nstype species.NamespaceType, _ string, result *Result) {
	andTasks := ""
	if result.Options.ScanTasks {
		andTasks = " and tasks"
	}
	if !result.Options.ScanProcs {
		log.Infof("skipping discovery of %s namespaces used by processes%s",
			andTasks, nstype.Name())
		return
	}
	log.Debugf("discovering %s namespaces used by processes%s...",
		andTasks, nstype.Name())

	// Things we want to do only once in order to not do them inside the
	// loops... yeah, we obviously don't trust the compiler to correctly
	// optimize this.
	nstypename := nstype.Name()
	nstypeidx := model.TypeIndex(nstype)
	hasForChildrenRef := nstype == species.CLONE_NEWPID || nstype == species.CLONE_NEWTIME
	var discoverOwnership determineNamespaceFlags
	if result.Options.DiscoverOwnership {
		discoverOwnership = detDiscoverOwnership
	}
	nsmap := result.Namespaces[nstypeidx]

	total := 0
	// For all processes (but not yet any tasks/threads) listed in /proc try to
	// gather the namespaces of a given type they use.
	//
	// Kindly reminder: result.Processes is a map, so Go will happily iterate it
	// in whatever random sequence it just takes a fancy of.
	for pid, proc := range result.Processes {
		// Discover the namespace instance of the specified type which this
		// particular process has joined. Please note that namespace
		// references for processes appear as symbolic(!) links in the /proc
		// filesystem, but in fact are behaving like hard links. Nevertheless,
		// we have to follow them like symbolic links in order to find the
		// identifier in form of the inode # of the referenced namespace.
		nsref := fmt.Sprintf("/proc/%d/ns/%s", pid, nstypename)
		foundns, isnew := determineNamespace(discoverOwnership,
			&proc.ProTaskCommon,
			nsref, nstype, nstypeidx, nsmap)
		if foundns == nil { // ...that didn't went well.
			continue
		}
		if isnew {
			log.Debugf("found namespace %s at %s[=%s]",
				foundns.(model.NamespaceStringer).TypeIDString(),
				nsref, proc.Name)
			total++
		}
		if !hasForChildrenRef {
			continue
		}
		foundns, isnew = determineNamespace(detForChildren|discoverOwnership,
			&proc.ProTaskCommon,
			nsref, nstype, nstypeidx,
			nsmap)
		if foundns == nil || !isnew {
			continue
		}
		log.Debugf("found namespace %s at %s[=%s]",
			foundns.(model.NamespaceStringer).TypeIDString(),
			nsref, proc.Name)
		total++
	}
	determineLeaders(nstype, result)
	// Try to set namespace references which we hope to be as long-lived as
	// possible; so we prefer the most senior leader process: the ealdorman.
	for _, ns := range nsmap {
		if ealdorman := ns.Ealdorman(); ealdorman != nil {
			ns.(namespaces.NamespaceConfigurer).SetRef(
				model.NamespaceRef{fmt.Sprintf("/proc/%d/ns/%s", ealdorman.PID, nstypename)})
		}
	}
	// Now scan tasks for yet unknown namespaces separately; this ensures that
	// for the already known namespaces we have also already established
	// process-based references that hopefully will be more stable than
	// task-based references.
	for _, proc := range result.Processes {
		procns := proc.Namespaces[nstypeidx]
		for _, task := range proc.Tasks {
			// tasks are actually also directly addressable in procfs using a
			// TID instead of a PID. They just don't show up when reading the
			// process directory.
			nsref := fmt.Sprintf("/proc/%d/ns/%s", task.TID, nstypename)
			newns, isnew := determineNamespace(detSetReference|discoverOwnership,
				&task.ProTaskCommon,
				nsref, nstype, nstypeidx, nsmap)
			if newns == nil {
				continue
			}
			if newns != procns {
				// tsk, tsk ... we've got a stray task here...
				newns.(namespaces.NamespaceConfigurer).AddLooseThread(task)
			}
			if isnew {
				log.Debugf("found namespace %s at %s[=task of %s]",
					newns.(model.NamespaceStringer).TypeIDString(),
					nsref, proc.Name)
				total++
			}
			if !hasForChildrenRef {
				continue
			}
			newns, isnew = determineNamespace(detSetReference|detForChildren|discoverOwnership,
				&task.ProTaskCommon,
				nsref, nstype, nstypeidx, nsmap)
			if newns == nil || !isnew {
				continue
			}
			if newns != procns {
				newns.(namespaces.NamespaceConfigurer).AddLooseThread(task)
			}
			log.Debugf("found namespace %s at %s[=%s]",
				newns.(model.NamespaceStringer).TypeIDString(),
				nsref, proc.Name)
			total++
		}
	}

	log.Infof("found %s joined by processes%s",
		andTasks, plural.Elements(total, "%s namespaces", nstype.Name()))
}

type determineNamespaceFlags uint8

const (
	detSetReference determineNamespaceFlags = 1 << iota
	detDiscoverOwnership
	detForChildren
)

// determineNamespace reads the details of the specified nsref namespace
// reference that must be of the specified type. It then updates the
// additionally specified namespace map. determineNamespace returns the
// discovered namespace, or nil if it could not be determined. The additional
// boolean is true if this is the first time this namespace is seen, otherwise
// false.
func determineNamespace(
	flags determineNamespaceFlags,
	procOrTask *model.ProTaskCommon,
	nsref string,
	nstype species.NamespaceType,
	nstypeidx model.NamespaceTypeIndex,
	nsmap model.NamespaceMap,
) (ns model.Namespace, firsttime bool) {
	if flags&detForChildren != 0 {
		nsref += "_for_children"
	}
	// Please note that we need the open fd further down below in case we need
	// to discover ownership.
	f, err := os.Open(nsref) // #nosec G304
	if err != nil {
		return nil, false
	}
	// Why not using a simple (typed) NamespacePath here? Because we want to
	// carry out multiple query operations and avoid repeated opening and
	// closing for each individual query on the same namespace.
	nsf, _ := ops.NewTypedNamespaceFile(f, nstype)
	defer nsf.Close() // ...we've taken over ownership of the *os.File as well!
	nsid, err := nsf.ID()
	if err != nil {
		return nil, false
	}
	ns, existingNs := nsmap[nsid]
	if !existingNs {
		// Only add a namespace we haven't yet seen. And yes, we don't give
		// a (file system-based) reference here, as we want to use a
		// reference from a leader process, and not of some child process
		// deep down the hierarchy, which might not even live for long (as
		// sad as this might be).
		var ref model.NamespaceRef
		if flags&detSetReference != 0 {
			ref = model.NamespaceRef{nsref}
		}
		ns = namespaces.New(nstype, nsid, ref)
		nsmap[nsid] = ns
	}
	// To speed up finding the process leaders in a specific namespace, we
	// remember this namespace as joined by the process we're just looking at.
	// Additionally, other applications also benefit from quickly navigating
	// from a process to its joined namespace proxy objects.
	//
	// However, if we're dealing with a PID or TIME namespace glanced from a
	// *_for_children reference, then we must not overwrite the (already) set
	// namespace the process itself is attached to. The *_for_children
	// references do not apply to the parent process, only to newly spawned
	// child processes.
	if flags&detForChildren == 0 {
		procOrTask.Namespaces[nstypeidx] = ns
	}
	// Let's also get the owning user namespace id, while we still have a
	// suitable fd open. For user namespaces, we skip this step, as this
	// is the same as the parent relationship. Additionally, it makes
	// things too awkward in the model, because then we would need to
	// treat ownership differently for non-user namespaces versus user
	// namespaces all the time. Thus, sorry, no user namespaces here.
	if flags&detDiscoverOwnership != 0 && nstype != species.CLONE_NEWUSER {
		ns.(namespaces.NamespaceConfigurer).DetectOwner(nsf)
	}
	return ns, !existingNs
}

// determineLeaders determines the leader processes for the discovered
// namespaces of a particular type of namespace, based on the process hiearchy.
func determineLeaders(nstype species.NamespaceType, result *Result) {
	nstypeidx := model.TypeIndex(nstype)
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
		leaderproc := proc
		parentproc := leaderproc.Parent
		for parentproc != nil && parentproc.Namespaces[nstypeidx] == leaderproc.Namespaces[nstypeidx] {
			leaderproc = parentproc
			parentproc = leaderproc.Parent
		}
		ns := leaderproc.Namespaces[nstypeidx]
		ns.(namespaces.NamespaceConfigurer).AddLeader(leaderproc)
	}
}
