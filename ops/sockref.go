// Copyright 2026 Harald Albrecht.
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

package ops

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/ioctl"
	"golang.org/x/sys/unix"
)

// openNamespacePath opens the namespace referenced by path and returns an
// O_CLOEXEC file descriptor to it. If path is a procfs fd link to a socket
// (“/proc/$PID/fd/$FD”), it returns the network namespace of that socket
// instead, as sockets cannot be opened via procfs.
func openNamespacePath(path string) (int, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if !errors.Is(err, unix.ENXIO) {
		return fd, err
	}
	if netnsfd, sockerr := socketNetnsFd(path); sockerr == nil {
		return netnsfd, nil
	}
	return -1, err
}

// openNamespaceFile is [openNamespacePath] returning an *[os.File] and errors
// in the same form as [os.Open].
func openNamespaceFile(path string) (*os.File, error) {
	fd, err := openNamespacePath(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(fd), path), nil
}

// socketNetnsFd returns a file descriptor to the network namespace of the
// socket referenced by the procfs fd link path, by temporarily duplicating the
// socket's fd into our process.
func socketNetnsFd(path string) (int, error) {
	pid, fdno, ok := parseProcFdPath(path)
	if !ok {
		return -1, unix.EINVAL
	}
	pidfd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return -1, err
	}
	defer func() { _ = unix.Close(pidfd) }()
	sockfd, err := unix.PidfdGetfd(pidfd, fdno, 0)
	if err != nil {
		return -1, err
	}
	defer func() { _ = unix.Close(sockfd) }()
	return ioctl.RetFd(sockfd, unix.SIOCGSKNS)
}

// parseProcFdPath returns the PID and fd number from a procfs fd link path of
// the form “.../$PID/fd/$FD” or “.../$PID/task/$TID/fd/$FD”.
func parseProcFdPath(path string) (pid int, fdno int, ok bool) {
	elems := strings.Split(path, "/")
	n := len(elems)
	if n < 4 || elems[n-2] != "fd" {
		return 0, 0, false
	}
	fd, err := strconv.ParseUint(elems[n-1], 10, 31)
	if err != nil {
		return 0, 0, false
	}
	pidelem := elems[n-3]
	if n >= 6 && elems[n-4] == "task" {
		if _, err := strconv.ParseUint(pidelem, 10, 31); err != nil {
			return 0, 0, false
		}
		pidelem = elems[n-5]
	}
	p, err := strconv.ParseUint(pidelem, 10, 31)
	if err != nil || p == 0 {
		return 0, 0, false
	}
	return int(p), int(fd), true
}
