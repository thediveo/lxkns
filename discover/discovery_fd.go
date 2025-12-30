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
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/ioctl"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// discoverFromFd discovers (1) namespaces from open file descriptors
// referencing namespaces either directly or instead sockets that are in turn
// attached to a network namespace, as well as (2) the socket-to-processes
// mapping in a single run. This way we avoid DRY of repeated open fd socket
// scanning.
//
// Please note that scanning file descriptors for namespaces and sockets
// automatically opts into discovering the socket-to-processes mapping, as this
// is a byproduct anyway.
//
// Since file descriptors are per process only, but not per task/thread, it
// sufficies to only iterate the process fd entries, leaving out the copies in
// the task fd entries.
func discoverFromFd(t species.NamespaceType, procfs string, result *Result) {
	if !result.Options.ScanFds && !result.Options.DiscoverSocketProcesses {
		slog.Info("skipping discovery of fd-referenced namespaces and socket processes")
		return
	}
	switch {
	case result.Options.ScanFds:
		slog.Debug("discovering fd-referenced namespaces and socket processes")
	default:
		slog.Debug("discovering socket processes")
	}
	scanFd(t, procfs, false, result)
}

const socketPrefix = "socket:["
const socketPrefixLen = len(socketPrefix)

// scanFd is discoverFromFd with special test harness handling enabled or
// disabled.
func scanFd(_ species.NamespaceType, procfs string, fakeprocfs bool, result *Result) {
	debugEnabled := slog.Default().Enabled(context.Background(), slog.LevelDebug)

	result.SocketProcessMap = SocketProcesses{}
	/* shorthand */ scanFds := result.Options.ScanFds
	// Iterate over all known processes, and then over all of their open file
	// descriptors. The /proc filesystem will give us the required
	// information.
	total := 0
	pidfd := 0 // grudingly accepting zero albeit being a valid fd, sigh.
	defer func() {
		// ensure to not leak a process fd in any case.
		if pidfd > 0 {
			_ = unix.Close(pidfd)
		}
	}()
	for pid := range result.Processes {
		// concatenating strings in combination with strconv.Itoa is roughly
		// 2.3× faster than to fmt.Sprintf, so its worth any minor inconvenience
		// anyway. (Intel Core i5 with amd64 architecture)
		basepath := procfs + "/" + strconv.Itoa(int(pid)) + "/fd"
		// avoid os.ReadDir as we don't want to waste CPU time on sorting the
		// directory entries.
		dirf, err := os.Open(basepath)
		if err != nil {
			continue
		}
		fdEntries, err := dirf.ReadDir(-1)
		_ = dirf.Close()
		if err != nil {
			continue
		}
		for _, fdEntry := range fdEntries {
			// Filter out all open file descriptors which are not symbolic
			// links; please note that there should only be symbolic links,
			// but better be careful here.
			if fdEntry.Type()&os.ModeSymlink == 0 {
				continue
			}
			// Let's read the link destination ("target") in order to get an
			// idea where it points to. We are interested in two variants out of
			// many more: first, those targets that name a type of namespace,
			// such as "net:[...]", and second, "socket:[...]" targets. The
			// socket targets can be queried for the network namespace the
			// socket is connected to.
			procFdPath := basepath + "/" + fdEntry.Name()
			fdDestination, err := os.Readlink(procFdPath)
			if err != nil {
				continue
			}
			var nsid species.NamespaceID
			var nstype species.NamespaceType
			var nsr relations.Relation
			if strings.HasPrefix(fdDestination, socketPrefix) {
				// It's a socket so we note down the relationship between the
				// socket's inode number and this process in any case, as this
				// is a byproduct of trying to find the socket's network
				// namespace.
				l := len(fdDestination)
				if l <= socketPrefixLen {
					continue
				}
				ino, err := strconv.ParseUint(fdDestination[8:l-1], 10, 64)
				if err != nil {
					continue
				}
				result.SocketProcessMap[ino] = append(result.SocketProcessMap[ino], pid)
				if !scanFds {
					continue
				}
				// So the calling explorer really wants to discover network
				// namespaces from sockets. If we haven't done yet for this
				// process, get a PID fd so we can later duplicate the
				// processes's fd into our process for further inspection.
				if pidfd <= 0 {
					pidfd, err = unix.PidfdOpen(int(pid), 0)
					if err != nil {
						continue
					}
				}
				nsid, nstype = namespaceOfSocket(pidfd, fdEntry.Name())
				if nstype == species.NaNS {
					continue
				}
			} else if !scanFds {
				// while scanning for sockets was requested, scanning fds
				// wasn't, so we then don't dig deeper into fds that might
				// reference namespaces directly.
				continue
			} else {
				nsid, nstype = namespaceFromLink(procFdPath, fdDestination, fakeprocfs)
				if nstype == species.NaNS {
					continue
				}
				nsr = ops.NamespacePath(procFdPath)
			}
			// Check if we already know this namespace, otherwise it's a new
			// discovery. Add such new discoveries and use the /proc fd path as
			// a path reference in case we want later to make use of this
			// namespace. Consumers of these /proc-based fd paths need to have a
			// clue about how to correctly deal with them in order to reference
			// the targeted namespace.
			nstypeidx := model.TypeIndex(nstype)
			if _, ok := result.Namespaces[nstypeidx][nsid]; ok {
				continue
			}
			foundns := namespaces.NewWithSimpleRef(nstype, nsid, procFdPath)
			if nsr != nil {
				foundns.(namespaces.NamespaceConfigurer).DetectOwner(nsr)
			}
			if debugEnabled {
				slog.Debug("found namespace",
					slog.String("namespace", foundns.(model.NamespaceStringer).TypeIDString()),
					slog.String("ref", procFdPath))
			}
			result.Namespaces[nstypeidx][nsid] = foundns
			total++
		}
		// Release the process fd as we don't need it anymore because we're
		// progressing to the next process in our list.
		if pidfd > 0 {
			_ = unix.Close(pidfd)
			pidfd = 0
		}
	}
	if scanFds {
		slog.Info("found namespaces",
			slog.String("src", "fd"), slog.Int("count", total))
	}
	slog.Info("found sockets", slog.Int("count", len(result.SocketProcessMap)))
}

// namespaceOfSocket returns the network namespace a particular socket fd (of
// the specified process) is connected to.
func namespaceOfSocket(pidfd int, fdname string) (species.NamespaceID, species.NamespaceType) {
	// PIDs are unsigned, but passed as int32...
	fdno, err := strconv.ParseUint(fdname, 10, 31)
	if err != nil {
		return species.NoneID, species.NaNS
	}

	// Duplicate the process' fd into our own process, then issue a query ioctl
	// on it to get the network namespace reference as another fd. This doesn't
	// mess with the other process' socket otherwise so that is safe to do:
	// look, but don't touch.
	sockfd, err := unix.PidfdGetfd(pidfd, int(fdno), 0)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	defer func() { _ = unix.Close(sockfd) }()
	netnsfd, err := ioctl.RetFd(sockfd, unix.SIOCGSKNS)
	if err != nil {
		return species.NoneID, species.NaNS
	}
	defer func() { _ = unix.Close(netnsfd) }()
	var netnsStat unix.Stat_t
	if err := unix.Fstat(netnsfd, &netnsStat); err != nil {
		return species.NoneID, species.NaNS
	}

	return species.NamespaceID{
		Dev: netnsStat.Dev,
		Ino: netnsStat.Ino,
	}, species.CLONE_NEWNET
}

// namespaceFromLink takes the target/destination (such as “net:[4026533429]”) a
// symbolic namespace link points to and returns its namespace ID (ino plus dev
// number) as well as the namespace type. In cased of error, it returns the type
// as species.NaNS.
//
// - path: "/proc/123/fd/42"
// - target: "net:[4026533429]"
func namespaceFromLink(path, target string, fakeprocfs bool) (
	species.NamespaceID, species.NamespaceType,
) {
	// Does the "symbolic" link point to a Linux kernel namespace?
	// This sorts out all other things, such as open sockets, et
	// cetera.
	nsid, nstype := species.IDwithType(target)
	if nstype == species.NaNS {
		return species.NoneID, species.NaNS
	}
	// ...remember that we want to follow the link and get the stat information
	// from where it points to; we don't want to get the stat for the fd (link)
	// entry itself. However, we can't stat the target path itself :(
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
