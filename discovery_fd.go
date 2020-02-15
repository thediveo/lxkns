// Discovers namespaces from the open file descriptors of process, which can
// be found in the /proc filesystem. Similar to the namespace discovery from
// processes, we need to run this discovery only once in the current PID
// namespace: the same reasoning again applies.

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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/thediveo/lxkns/nstypes"
)

// discoverFromFd discovers namespaces from process file descriptors
// referencing them. Since file descriptors are per process only, but not per
// task/thread, it sufficies to only iterate the process fd entries, leaving
// out the copies in the task fd entries.
func discoverFromFd(_ nstypes.NamespaceType, procfs string, result *DiscoveryResult) {
	if result.Options.SkipFds {
		return
	}
	// Iterate over all known processes, and then over all of their open file
	// descriptors. The /proc filesystem will give us the required
	// information.
	for pid := range result.Processes {
		basepath := fmt.Sprintf(filepath.Join(procfs, "%d/fd"), pid)
		fdentries, err := ioutil.ReadDir(basepath)
		if err != nil {
			continue
		}
		for _, fdentry := range fdentries {
			// Filter out all open file descriptors which are not symbolic
			// links; please note that there should only be symbolic links,
			// but better be careful here.
			if fdentry.Mode()&os.ModeSymlink == 0 {
				continue
			}
			target, err := os.Readlink(basepath + "/" + fdentry.Name())
			if err != nil {
				continue
			}
			// Does the "symbolic" link point to a Linux kernel namespace?
			// This sorts out all other things, such as open sockets, et
			// cetera.
			nsid, nstype := nstypes.IDwithType(target)
			if nstype == nstypes.NaNS {
				continue
			}
			// Check if we already know this namespace, otherwise is a new
			// discovery. Add such new discoveries and use the /proc fd path
			// as a path reference in case we want later to make use of this
			// namespace.
			nstypeidx := TypeIndex(nstype)
			if _, ok := result.Namespaces[nstypeidx][nsid]; ok {
				continue
			}
			result.Namespaces[nstypeidx][nsid] = NewNamespace(
				nstype, nsid, basepath+"/"+fdentry.Name())
		}
	}
}
