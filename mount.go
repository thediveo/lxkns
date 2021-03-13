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

package lxkns

import (
	"io"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
)

// MountEnterNamespaces takes a mount namespace to be entered and returns the
// namespaces which need to be specified to ReexecIntoAction. It takes into
// account if first entering the correct user namespace might allow to enter
// the desired mount namespace even without having the necessary effective
// capabilities right now, but we could gain them by entering the user
// namespace first.
func MountEnterNamespaces(
	mntns model.Namespace, namespaces model.AllNamespaces) []model.Namespace {
	// If we're running without the necessary privileges to change into mount
	// namespaces, but we are running under the user which is the owner of the
	// mount namespace, then we first gain the necessary privileges by
	// switching into the user namespace for the mount namespace we're the
	// owner (creator) of, and then can successfully enter the mount
	// namespaces. And yes, this is how Linux namespaces, and especially the
	// user namespaces and setns() are supposed to work.
	ownusernsid, _ := ops.NamespacePath("/proc/self/ns/user").ID()
	enterns := []model.Namespace{mntns}
	if usermntnsref, err := ops.NamespacePath(mntns.Ref()).User(); err == nil {
		usernsid, _ := usermntnsref.ID()
		// Do not leak, release user namespace immediately, as we're done with
		// it.
		_ = usermntnsref.(io.Closer).Close()
		if userns, ok := namespaces[model.UserNS][usernsid]; ok &&
			userns.ID() != ownusernsid {
			// Prepend the user namespace to the list of namespaces we need to
			// enter, due to the magic capabilities of entering user
			// namespaces. And, by the way, worst programming language syntax
			// ever, even more so than Perl. TECO isn't in the competition,
			// though.
			enterns = append([]model.Namespace{userns}, enterns...)
		}
	}
	return enterns
}
