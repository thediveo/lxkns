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

//go:build linux

package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thediveo/ioctl"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/plural"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// discoverFromFd discovers namespaces from process file descriptors referencing
// them either directly or via socket fds. Since file descriptors are per
// process only, but not per task/thread, it sufficies to only iterate the
// process fd entries, leaving out the copies in the task fd entries.
func discoverFromFd(t species.NamespaceType, procfs string, result *Result) {
	if !result.Options.ScanFds {
		log.Infof("skipping discovery of fd-referenced namespaces")
		return
	}
	log.Debugf("discovering fd-referenced namespaces...")
	scanFd(t, procfs, false, result)
}

// namespaceFromFd is discoverFromFd with special test harness handling enabled
// or disabled.
func scanFd(_ species.NamespaceType, procfs string, fakeprocfs bool, result *Result) {
	// Iterate over all known processes, and then over all of their open file
	// descriptors. The /proc filesystem will give us the required
	// information.
	total := 0
	for pid := range result.Processes {
		basepath := fmt.Sprintf(filepath.Join(procfs, "%d/fd"), pid)
		fdentries, err := os.ReadDir(basepath)
		if err != nil {
			continue
		}
		for _, fdentry := range fdentries {
			// Filter out all open file descriptors which are not symbolic
			// links; please note that there should only be symbolic links,
			// but better be careful here.
			if fdentry.Type()&os.ModeSymlink == 0 {
				continue
			}
			// Let's read the link destination ("target") in order to get an
			// idea where it points to. We are interested in two variants out of
			// many more: first, those targets that name a type of namespace,
			// such as "net:[...]", and second, "socket:[...]" targets. The
			// socket targets can be queried for the network namespace the
			// socket is connected to.
			path := basepath + "/" + fdentry.Name()
			target, err := os.Readlink(path)
			if err != nil {
				continue
			}
			var nsid species.NamespaceID
			var nstype species.NamespaceType
			if strings.HasPrefix(target, "socket:[") {
				nsid, nstype = namespaceSocket(pid, fdentry.Name())
			} else {
				nsid, nstype = namespaceLink(path, target, fakeprocfs)
			}
			if nstype == species.NaNS {
				continue
			}
			// Check if we already know this namespace, otherwise is a new
			// discovery. Add such new discoveries and use the /proc fd path
			// as a path reference in case we want later to make use of this
			// namespace.
			nstypeidx := model.TypeIndex(nstype)
			if _, ok := result.Namespaces[nstypeidx][nsid]; ok {
				continue
			}
			log.Debugf("found namespace %s:[%d] at %s", nstype.Name(), nsid.Ino, path)
			result.Namespaces[nstypeidx][nsid] = namespaces.NewWithSimpleRef(
				nstype, nsid, basepath+"/"+fdentry.Name())
			total++
		}
	}
	log.Infof("discovered %s", plural.Elements(total, "fd-referenced namespaces"))
}

// namespaceSocket returns the network namespace a socket fd is connected to.
func namespaceSocket(pid model.PIDType, fdname string) (species.NamespaceID, species.NamespaceType) {
	fdno, err := strconv.ParseUint(fdname, 10, 32)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	pidfd, err := unix.PidfdOpen(int(pid), 0)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	defer unix.Close(pidfd)
	sockfd, err := unix.PidfdGetfd(pidfd, int(fdno), 0)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	defer unix.Close(sockfd)
	netnsfd, err := ioctl.RetFd(sockfd, unix.SIOCGSKNS)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	defer unix.Close(netnsfd)
	var netnsStat unix.Stat_t
	if err := unix.Fstat(netnsfd, &netnsStat); err != nil {
		return species.NoneID, species.NaNS
	}
	return species.NamespaceID{
		Dev: netnsStat.Dev,
		Ino: netnsStat.Ino,
	}, species.CLONE_NEWNET
}

// namespaceLink takes the target/destination a symbolic namespace link points
// to and returns its namespace ID (ino plus dev number) as well as the
// namespace type. It returns a type of species.NaNS in case of any error.
func namespaceLink(path, target string, fakeprocfs bool) (species.NamespaceID, species.NamespaceType) {
	// Does the "symbolic" link point to a Linux kernel namespace?
	// This sorts out all other things, such as open sockets, et
	// cetera.
	nsid, nstype := species.IDwithType(target)
	if nstype == species.NaNS {
		return species.NoneID, species.NaNS
	}
	// ...remember that we want to follow the link and get the stat information
	// from where it points to; we don't want to get the stat for the fd entry
	// itself. However, we can't stat the target path itself :(
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		if !fakeprocfs {
			return species.NoneID, species.NaNS
		}
		if err := unix.Lstat(path, &stat); err != nil {
			return species.NoneID, species.NaNS
		}
	}
	nsid.Dev = stat.Dev
	return nsid, nstype
}
