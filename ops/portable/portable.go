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

package portable

import (
	"fmt"
	"os"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// PortableReference describes one or several aspects of a portable namespace
// reference, where portable refers to being transferable between different
// processes on the same host, but not across different hosts.
type PortableReference struct {
	ID        species.NamespaceID   // if known, ID of namespace
	Type      species.NamespaceType // if known, type of namespace
	Path      string                // if known, filesystem path reference to namespace
	PID       model.PIDType         // if known, process for verification
	Starttime uint64                // if known, start time of process for verification
}

// Open opens a portable namespace reference and validates the opened namespace
// against the information contained in the portable reference.
//
// When the caller supplies a namespace ID instead of a namespace path, then
// Open will try to locate a suitable namespace path first by running a
// namespace discovery and trying to find the namespace by its ID, using the
// path of a match. (See also: LocateNamespace)
//
// When the caller supplies a path and if the namespace referenced by this path
// doesn't match an optionally specified ID and/or type, then Open fails.
//
// When opening a portable namespace reference, the following checks are carried
// out (or not), depending on the information supplied:
//
//   - just the namespace ID is known, nothing else: namespace is located via ID requiring a discovery.
//   - namespace ID and type is known: namespace is located ID with discovery, then type is cross-checked.
//   - just the path is known, nothing else: no checks beyond successfully opening as a namespace.
//   - path and type is known: cross-checking the type against what has been opened.
//   - path and ID are known: path is used and checked against the ID.
//   - path, ID and type are known: path is used and checked against ID and type of namespace.
//   - (path OR ID) is known, as well as PID and Starttime: additionally checks that process is still around.
func (portref PortableReference) Open() (rel relations.Relation, closer func(), err error) {
	// If we don't have a path, we must try to locate the namespace through a
	// discovery and then looking for the namespace ID...
	path := portref.Path
	if path == "" {
		ns := LocateNamespace(portref.ID, portref.Type)
		if ns == nil {
			return nil, nil, fmt.Errorf("cannot locate namespace %s:[%d]",
				portref.Type.Name(), portref.ID.Ino)
		}
		if len(ns.Ref()) != 1 {
			return nil, nil, fmt.Errorf("invalid multi-element reference %s", ns.Ref().String())
		}
		path = ns.Ref()[0]
	}
	return portref.openPath(path)
}

// openPath does the heavy lifting of opening the namespace, specified by its
// path.
func (portref PortableReference) openPath(path string) (rel relations.Relation, closer func(), err error) {
	// Before we run further checks, we have to open the namespace first, so
	// that it cannot get destroyed anymore as long as we keep the fd open that
	// is referencing it.
	nsf, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, nil, err
	}
	// Ensure to properly close the open namespace reference in case we fail
	// down below somewhere while deep in the cross-checks.
	defer func() {
		if err != nil { // here we already know that nsf won't be nil.
			_ = nsf.Close()
		}
	}()
	// Optionally check that we have opened the correct type of namespace.
	var nstype species.NamespaceType // TODO: explicit type?
	if portref.Type != 0 {
		if nstype, err = ops.NamespacePath(path).Type(); err != nil || portref.Type != nstype {
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, fmt.Errorf(
				"portable namespace reference type mismatch, expected %s, got %s",
				portref.Type.Name(), nstype.Name())
		}
	}
	typedref, err := ops.NewTypedNamespaceFile(nsf, nstype)
	if err != nil { // unlikely to fail at all, as we already did the type check.
		return nil, nil, err
	}
	// Optionally check that we've got the right namespace instance.
	if portref.ID != species.NoneID {
		if realid, err := typedref.ID(); err != nil || realid != portref.ID {
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, fmt.Errorf(
				"portable namespace reference ID mismatch, expected %s:[%d], got %[1]s:[%[3]d]",
				nstype.Name(), portref.ID.Ino, realid.Ino)
		}
	}
	// Optionally check for the correct (associated) process; we cannot do this
	// before we've opened the namespace reference, as otherwise the check would
	// end up with a race condition on the bad side: between process check and
	// opening. And we're slightly paranoid in that we even cross-check that the
	// specified process is still attached to the namespace in question. Not
	// that is completely bullet-proof, but it raises the bar.
	if portref.PID != 0 {
		proc := model.NewProcess(portref.PID, false)
		if proc == nil || proc.Starttime != portref.Starttime {
			return nil, nil, fmt.Errorf("process PID %d is gone", portref.PID)
		}
	}
	// Okay, we successfully passed the checks, so the caller is now entitled to
	// use the opened namespace reference...
	return typedref, func() { _ = nsf.Close() }, nil
}
