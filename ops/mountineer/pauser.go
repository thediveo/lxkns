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

import "github.com/thediveo/lxkns/model"

// Pauser keeps a mount namespace accessible via the process filesystem. Any
// Pauser returned by some constructor/factory must be ready-to-use. In
// particular, the mount namespace must have been successfully opened and
// accessible through the pauser's process or task process filesystem entry.
type Pauser interface {
	// PID (or TID) of a process or task that can be used to access a mount
	// namespace via the process filesystem.
	PID() model.PIDType

	// Closes the Pauser (by terminating it) and releases allocated system
	// resources. Close is idempotent.
	Close()
}
