// Copyright 2021 Harald Albrecht.
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

package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

// scanCgroups scans all processes for their control groups; it scans only on a
// specific type of controller, the "cpu" v1 controller on (1) the assumption
// that this controller is widely used and (2) we're interested in the fridge
// (well, "freezer") state. On a side note, the "memory" controller
// unfortunately has been disabled on some architectures (ARM) for quite some
// time.
func (p ProcessTable) scanCgroups() {
	for pid, proc := range p {
		controllers := processCgroup(cgrouptypes, pid)
		proc.CpuCgroup = controllers[0]
		proc.FridgeCgroup = controllers[1]
	}
}

var cgrouptypes = []string{"cpu", "freezer"}

// scanFridges discovers the freezer states in the cgroups hierarchy; either
// from the cgroups v1 freezer hierarchy if available, or from the unified
// cgroups v2 hierarchy.
func (p ProcessTable) scanFridges() {
	// First determine the list of unique cgroup freezer paths, as not every
	// process will have its own personal freezer and we can thus avoid reading
	// the same states over and over again as well as sending the same freezer
	// path multiple times to the potentially re-executed action.
	fridges := map[string]bool{} // maps unique paths to freezer states
	for _, proc := range p {
		fridges[proc.FridgeCgroup] = false
	}
	fridgepaths := make([]string, len(fridges)) // our unique list to be.
	idx := 0
	for fridgepath := range fridges {
		fridgepaths[idx] = fridgepath
		idx++
	}
	// Now read the freezer states, but switch into the initial mount namespace
	// first if necessary...
	var frozens []bool
	initialns, ierr := ops.NewTypedNamespacePath("/proc/1/ns/mnt", species.CLONE_NEWNS).ID()
	myns, myerr := ops.NewTypedNamespacePath("/proc/self/ns/mnt", species.CLONE_NEWNS).ID()
	if ierr == nil && myerr == nil && myns != initialns {
		// Nota bene: no shenangiens possible and thus necessary when it comes
		// to user namespaces: either we're in the same initial user namespace
		// or it is game over anyway.
		if err := reexec.RunReexecAction(
			"discover-fridges",
			reexec.Namespaces([]reexec.Namespace{{Type: "mnt", Path: "/proc/1/ns/mnt"}}),
			reexec.Result(&frozens),
		); err != nil {
			log.Errorf("could not determine fridge states in mnt:[%d]: %s", initialns.Ino, err.Error())
		}
	} else {
		frozens = fridgeStates(fridgepaths)
	}
	// ...and finally distribute the freezer states into the appropriate process
	// objects. As there is no stable iteration order over the map between
	// cgroup paths and freezer states we now propagate the states into the map,
	// so we finally can look up these states based on path. Remember that we
	// limited the amount of data transferred forth and back the re-executed
	// action as much as possible.
	for idx, fridgepath := range fridgepaths {
		fridges[fridgepath] = frozens[idx]
	}
	for _, proc := range p {
		proc.FridgeFrozen = fridges[proc.FridgeCgroup]
	}
}

// Register discoverFridges() as an action for re-execution.
func init() {
	reexec.Register("discover-fridges", discoverFridges)
}

// discoverFridges is the reexec action run in a separate mount namespace (in
// particular, the initial mount namespace) in order to gather the freezer
// states of all cgroup "fridges".
func discoverFridges() {
	var fridgepaths []string
	if err := json.NewDecoder(os.Stdin).Decode(&fridgepaths); err != nil {
		panic(err.Error())
	}
	if err := json.NewEncoder(os.Stdout).Encode(fridgeStates(fridgepaths)); err != nil {
		panic(err.Error())
	}
}

// fridgeStates determines the (effective) freezer states for the cgroup paths
// specified. It needs to be run in the initial mount namespace in order to have
// full view on the cgroup freezer hierarchy, as otherwise the freezer states
// cannot be determined. The cgroup freezer paths are relative to the
// auto-discovered cgroups hierarchy root, albeit usually specified as (pseudo)
// absolute hierarchy paths (due to some ancient Linux kernel penguin foo).
func fridgeStates(fridgepaths []string) (frozens []bool) {
	fridgeroot, unified := fridgeRoot()
	frozens = make([]bool, len(fridgepaths))
	if unified {
		// ...me not trusting Golang's toolchain to correctly optimizing the
		// constant check out of the loop...
		for idx, fridgepath := range fridgepaths {
			frozens[idx] = frozenV2(filepath.Join(fridgeroot, fridgepath))
		}
	} else {
		for idx, fridgepath := range fridgepaths {
			frozens[idx] = frozenV1(filepath.Join(fridgeroot, fridgepath))
		}
	}
	return
}

// frozenV2 returns the cgroups v2 freezer effective status information for the
// specified process.
func frozenV2(fridgepath string) (frozen bool) {
	// But, where's v1's "freezer.state", that is, the process' effective state?
	// It can now be found as one of possibly many event entries in
	// "cgroup.events". Yuk. But please also note that the v2 root cgroup
	// doesn't have the "cgroup.freeze" interface file. Other than that, see
	// https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#core-interface-files
	// for details.
	if events, err := ioutil.ReadFile(filepath.Join(fridgepath, "cgroup.events")); err == nil {
		for _, event := range strings.Split(string(events), "\n") {
			if strings.HasPrefix(event, "frozen ") {
				if event[7] == '1' {
					frozen = true
				}
				break
			}
		}
	}
	return
}

// frozenV1 returns the effective freezer status information for the specified
// process and maps it onto our v2-like simplified state model.
func frozenV1(fridgepath string) (frozen bool) {
	// Please note: "the root cgroup is non-freezable and the above
	// interface files don't exist."
	// (https://www.kernel.org/doc/Documentation/admin-guide/cgroup-v1/freezer-subsystem.rst)
	if state, err := ioutil.ReadFile(
		filepath.Join(fridgepath, "freezer.state")); err == nil {
		switch strings.TrimSuffix(string(state), "\n") {
		case "FREEZING":
			fallthrough
		case "FROZEN":
			frozen = true
		}
	} else {
		log.Warnf("cannot access freezer state %q, err: %s", fridgepath, err.Error())
	}
	return
}

// processCgroup returns the name (hierarchy path) of some of the cgroup
// controllers a specific process is in (as specified in the controllertypes
// parameter).
//
// We first try to find the specified cgroup v1 controllers if available and
// only then fall back to the unified cgroups v2 hierarchy.
//
// Note: the cgroup path(s) returned is (are) relative to the cgroups roots in
// the mount namespace of the **current** process, even as they start with "/"
// (at least when discovered inside the initial pid+cgroup namespaces).
func processCgroup(controllertypes []string, pid PIDType) (paths []string) {
	paths = make([]string, len(controllertypes))
	cgroup, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return
	}
	defer cgroup.Close()
	scanner := bufio.NewScanner(cgroup)
	unifiedroot := "" // (if detected) the cgroups v2 unified hierarchy root
	for scanner.Scan() {
		if err == nil {
			// See https://man7.org/linux/man-pages/man7/cgroups.7.html, section
			// "NOTES", subsection "/proc files". For cgroups v1 controllers,
			// the second field specifies the comma-separated list of the
			// controllers bound to the hierarchy: here, we look for, say, the
			// "cpu" controller. The third field specifies the path in the
			// cgroups hierarchy; it is relative to the mount point of the
			// hierarchy -- which in turn depends on the mount namespace of this
			// process :)
			//
			// For the unified cgroups v2 hierarchy the second field will be
			// empty, which otherwise would specify the particular cgroup v1
			// hierarchy/-ies.
			if fields := strings.Split(scanner.Text(), ":"); len(fields) == 3 {
				if fields[1] != "" {
					// cgroups v1 hierarchies
					controllers := strings.Split(fields[1], ",")
					for _, ctrl := range controllers {
						for idx, controllertype := range controllertypes {
							if ctrl == controllertype {
								paths[idx] = fields[2]
							}
						}
					}
				} else {
					// when we come across a single unified cgroups v2 hierarchy
					// root, remember it so we can later fix any missing
					// controller paths.
					unifiedroot = fields[2]
				}
			}
		}
	}
	// Now fix the missing cgroups controller paths we couldn't satisfy from v1
	// (if present) using the unified v2 hierarchy path. We're simplifying here
	// and don't look for a specific controller type (which we might find up
	// only higher up the hierarchy) but instead just take the unified path,
	// basta.
	for idx, path := range paths {
		if path == "" {
			paths[idx] = unifiedroot
		}
	}
	// Hopefully, we've gathered all controller paths by now.
	return
}

// fridgeRoot determines the root of the cgroup freezer hierarchy (in the
// current mount namespace). For cgroups v1 this usually is its own hierarchy,
// for cgroups v2 this is the unified hierarchy.
func fridgeRoot() (root string, unified bool) {
	// First search for a cgroups v1 freezer hierarchy, because a v2 unified
	// hierarchy might also be present ... now don't let us worry about some
	// software already using the v2 freezer in a hybrid configuration (argh).
	for _, mountinfo := range mntinfo.MountsOfType(-1, "cgroup") {
		for _, sopt := range strings.Split(mountinfo.SuperOptions, ",") {
			if sopt == "freezer" {
				root = mountinfo.MountPoint
				return
			}
		}
	}
	// ...otherwise, there must be a cgroups v2 unified hierarchy.
	mountinfo := mntinfo.MountsOfType(-1, "cgroup2")
	if len(mountinfo) > 0 {
		unified = true
		root = mountinfo[0].MountPoint
	}
	return
}
