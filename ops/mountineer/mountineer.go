// Copyright 2021 Harald Albrecht.
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

package mountineer

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/procfsroot"
)

// Mountineer takes a namespace reference, where this namespace reference might
// even be located in some other mount namespace (to be reached by a series of
// mount namespace references).
type Mountineer struct {
	// Target mount namespace reference, optionally preceded by contextual mount
	// namespace references. The first reference is always taken in the context
	// of the initial mount namespace, each following reference then in the
	// context of its preceding mount namespace.
	ref model.NamespaceRef
	// root path for addressing paths and directories ("contents") in the file
	// system view provided by a mount namespace.
	contentsRoot string
	// pause/sandbox process, if any.
	sandbox Pauser
	// PID to report back: this can be either of the pause/sandbox process or
	// PID 1 (alternatively the own PID if PID 1 is out of reach), depending on
	// the reference: this allows API users to unify usage without having to
	// differentiate between having a sandbox and PID or not when dealing with
	// existing Go modules working on /proc nodes.
	pid model.PIDType
}

// Check whether PID 1 (more precisely, /proc/1) is accessible to us or fall
// back to our own PID. This is mostly useful to users of the CLI tools.
func init() {
	if _, err := os.Stat("/proc/1/ns/mnt"); err != nil {
		initialContextPID = model.PIDType(os.Getpid())
	}
}

// initialContextPID is the PID to use as the initial context for resolving
// non-proc reference path elements.
var initialContextPID = model.PIDType(1)

// NewWithMountNamespace is like New, but instead expects a mount namespace
// interface as opposed to a namespace reference in form of one or more VFS
// paths. It only accepts mount namespaces, not any other type of namespace. It
// optimizes the case where the mount namespace has a process attached.
func NewWithMountNamespace(mountns model.Namespace, usernsmap model.NamespaceMap) (*Mountineer, error) {
	if mountns.Type() != species.CLONE_NEWNS {
		return nil, errors.New("invalid non-mount namespace " +
			mountns.(model.NamespaceStringer).TypeIDString())
	}
	if ealdorman := mountns.Ealdorman(); ealdorman != nil {
		return &Mountineer{
			ref:          mountns.Ref(),
			contentsRoot: "/proc/" + strconv.FormatUint(uint64(ealdorman.PID), 10) + "/root",
			pid:          ealdorman.PID,
		}, nil
	}
	return New(mountns.Ref(), usernsmap)
}

// New opens the mount namespace for file access and returns a new managing
// Mountineer. The mount namespace is referenced in one of the following ways:
//
// A. a single mount namespace reference that can be addressed in the initial
// mount namespace. If this reference isn't inside the /proc file system, then
// it will automatically be taken as relative to the initial mount namespace,
// that is, the mount namespace of process PID 1.
//
// B. a sequence of mount namespace references that need to be opened first
// before the final namespace reference.
//
// Specifying a map of user namespaces allows entering mount namespaces in those
// situations where the caller has insufficient capabilities itself but has
// sufficient capabilities in the user namespace owning the mount namespace.
func New(ref model.NamespaceRef, usernsmap model.NamespaceMap) (m *Mountineer, err error) {
	if len(ref) == 0 {
		return nil, errors.New("cannot open zero mount namespace reference")
	}
	// Make sure to always clean up in case of an error on our way out. This is
	// probably one of the rare cases where named return values might do some
	// good.
	m = &Mountineer{ref: ref}
	defer func() {
		if err != nil {
			m.Close()
			m = nil // ensure to never return a mountineer in case of error.
		}
	}()
	// The starting context for mount namespace references is the initial mount
	// namespace (if it's not in the procfs anyway).
	pid := model.PIDType(0) // sic! we won't want to kill the init process ;)
	// Now work along the list of mount namespace references, switching contexts
	// along the way as we make progress...
	for idx, refpath := range ref {
		// Sanity check: empty and non-absolute reference paths are considered
		// invalid.
		if refpath == "" || refpath[0] != '/' {
			err = errors.New("invalid mount namespace reference " + ref[:idx+1].String())
			return
		}
		// Sanity check: only a single (=first) reference is allowed to
		// reference the proc filesystem. Otherwise, we consider any stray proc
		// filesystem reference to be a violation.
		if strings.HasPrefix(refpath, "/proc/") {
			if idx != 0 {
				err = errors.New("invalid mount namespace " + ref[:idx+1].String() +
					" reference in multi-ref context")
				return
			}
			// This is a (single) proc filesystem reference that we can directly
			// reference without any need for pause processes.
			//
			//      / proc / $PID / root / ...
			//  [0]   [1]    [2]   [3]     [4]...
			r := strings.SplitN(refpath, "/", 4)
			if len(r) < 4 || r[3] == "" {
				err = errors.New("invalid mount namespace reference " + ref[:idx+1].String())
				return
			}
			var rooterpid int64
			rooterpid, err = strconv.ParseInt(r[2], 10, 32)
			if err != nil || rooterpid <= 0 {
				err = errors.New("invalid mount namespace reference " + ref[:idx+1].String())
				return
			}
			pid = model.PIDType(rooterpid)
			m.pid = pid
			m.contentsRoot = "/proc/" + r[2] + "/root"
			continue
		}
		// It's a bind-mounted mount namespace reference, to be taken in the
		// current context. The current context is a process, either the initial
		// process (to get things going) or a pause process.
		rooterpid := uint64(pid)
		if rooterpid == 0 {
			rooterpid = uint64(initialContextPID) // initial context is mount namespace of init process.
		}
		contentsRoot := "/proc/" + strconv.FormatUint(rooterpid, 10) + "/root"
		var evilrefpath string
		evilrefpath, err = procfsroot.EvalSymlinks(refpath, contentsRoot, procfsroot.EvalFullPath)
		if err != nil {
			err = errors.New("invalid mount namespace reference " + ref[:idx+1].String() + ":" + err.Error())
			return
		}
		// Start a pause process and attach it to the mount namespace referenced
		// by "ref" (in the mount namespace reachable via process with "pid").
		wormholedref := contentsRoot + evilrefpath
		log.Debugf("opening pandora's sandbox at %s", wormholedref)
		var sandbox Pauser
		if usernsref := m.usernsref(wormholedref, usernsmap); usernsref != "" {
			sandbox, err = newPauseProcess(wormholedref, usernsref)
		} else {
			sandbox, err = newPauseTask(wormholedref)
		}
		if err != nil {
			err = fmt.Errorf("sandbox failure, reason: %w", err)
			return
		}
		// The sandbox has attached to the mount namespace, now we can "safely"
		// access the latter via the proc file system.
		pid = sandbox.PID()
		// Retire the previous sandbox, if any. Do not retire the process giving
		// us the initial context though ... that would be ... bad, really bad.
		m.Close()
		m.pid = pid
		m.sandbox = sandbox // switch over, and rinse+repeat.
		m.contentsRoot = "/proc/" + strconv.FormatUint(uint64(pid), 10) + "/root"
	}
	// ...and we keep the last sandbox open.
	return m, nil
}

// Close the network namespace that previously has been "opened" by this
// Mountneneer, releasing any additional resources that might have been needed
// for opening the mount namespace and keeping it open.
func (m *Mountineer) Close() {
	if m.sandbox != nil {
		m.sandbox.Close()
		m.sandbox = nil
	}
}

// Ref returns the mount namespace reference.
func (m *Mountineer) Ref() model.NamespaceRef {
	return m.ref
}

// Open opens the named file for reading, resolving the specified name correctly
// for any symbolic links in the context of the particular mount namespace.
func (m *Mountineer) Open(name string) (*os.File, error) {
	return m.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile opens the named file with the specified flag, using the mode perm
// when creating new files. The specified name is resolved correctly for any
// symbolic links in the context of the particular mount namespace.
func (m *Mountineer) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	pathname, err := m.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(pathname, flag, perm) // #nosec G304
}

// ReadFile reads all contents of the named file, returning it as a byte slice.
func (m *Mountineer) ReadFile(name string) ([]byte, error) {
	pathname, err := m.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(pathname) // #nosec G304
}

// ReadDir reads the named directory, returning all its directory entries sorted
// by filename. If there was an error, then ReadDir returns not only the error,
// but for fun also all directory entries it was able to read up to the point of
// the error.
func (m *Mountineer) ReadDir(name string) ([]fs.DirEntry, error) {
	pathname, err := m.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.ReadDir(pathname)
}

// Resolve resolves a pathname inside the open mount namespace to a pathname
// that can be used by a caller in a different mount namespace, using a
// host-wide PID view. If the specified pathname is not absolute it is taken
// relative to the current working directory.
func (m *Mountineer) Resolve(pathname string) (string, error) {
	var err error
	pathname, err = filepath.Abs(pathname)
	if err != nil {
		return "", err
	}
	// If we don't need to use a wormhole to a different mount namespace then
	// simply return the absolute path as is.
	if m.contentsRoot == "" {
		return pathname, nil
	}
	// EvalSymlinks returns the evaluated paths always as absolute, but without
	// the separate (prefixing) root.
	pathname, err = procfsroot.EvalSymlinks(pathname, m.contentsRoot, procfsroot.EvalFullPath)
	if err != nil {
		return "", err
	}
	return m.contentsRoot + pathname, nil
}

// PID returns the PID of the sandbox pauser (if any), or PID 1 in case a
// sandbox wasn't needed.
func (m *Mountineer) PID() model.PIDType {
	return m.pid
}

// usernsref returns a reference to the user namespace owning a mount namespace
// if the user namespace of the current process and the user namespace of the
// mount namespace differ. Otherwise, it returns an empty reference "",
// indicating that user namespace switching isn't necessary.
//
// Please note that usernsref does not check if switching the user namespace
// will actually be possible.
func (m *Mountineer) usernsref(mntnsref string, usernsmap model.NamespaceMap) string {
	if usernsmap == nil {
		// Without a user namespace map we cannot determine a user namespace
		// reference in case switching the user namespace is necessary to enter
		// the mount namespace.
		return ""
	}
	// If we're running without the necessary privileges to change into mount
	// namespaces, but we are running under the user which is the owner of the
	// mount namespace, then we first gain the necessary privileges by switching
	// into the user namespace for the mount namespace we're the owner (creator)
	// of, and then can successfully enter the mount namespaces. And yes, this
	// is how Linux namespaces are supposed to work, and especially the user
	// namespaces and setns().
	ownusernsid, _ := ops.NamespacePath("/proc/self/ns/user").ID()
	mntuserns, err := ops.NamespacePath(mntnsref).User()
	if err != nil {
		return ""
	}
	mntusernsid, _ := mntuserns.ID()
	_ = mntuserns.(io.Closer).Close() // ...do not leak.
	if mntusernsid == ownusernsid {
		return "" // same owning user namespace, no need to switch userns.
	}
	// So we want to try to switch into the user namespace owning the mount
	// namespace first. The complication here is: the Linux kernel gave us a
	// file descriptor and that tells us what the ID of that user namespace is.
	// But it doesn't tell us how to address that user namespace, bummer.
	userns, ok := usernsmap[mntusernsid]
	if !ok {
		return ""
	}
	if len(userns.Ref()) != 1 {
		// We do not support bind-mounted owning user namespaces, as this would
		// make switching the sandbox process correctly very difficult to
		// achieve. It's much easier to simply run the API caller in the initial
		// user namespace, so it has access to the namespaces belonging to child
		// user namespaces.
		return ""
	}
	return userns.Ref()[0]
}
