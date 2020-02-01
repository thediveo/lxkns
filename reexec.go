// Simple wrapper for reexec.ForkReexec() which accepts lxkns Namespaces.

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

import "github.com/thediveo/lxkns/reexec"

// ReexecIntoAction forks and then re-executes this process in order to run a
// specific action (indicated by actionname) in a set of (different) Linux
// kernel namespaces. The stdout result of running the action is then
// deserialized as JSON into the specified result element.
func ReexecIntoAction(actionname string, namespaces []Namespace, result interface{}) (err error) {
	rexns := make([]reexec.Namespace, len(namespaces))
	for idx := range namespaces {
		rexns[idx].Type = namespaces[idx].Type().String()
		rexns[idx].Path = namespaces[idx].Ref()
	}
	return reexec.ForkReexec(actionname, rexns, result)
}

// HandleDiscoveryInProgress must be called from an application's main()
// function as early as possible. It checks if the current process is an
// action invocation: if this is the case, the requested action is called, and
// the process then terminated. HandleDiscoveryInProgress only returns if
// there is no action to be taken.
func HandleDiscoveryInProgress() {
	reexec.CheckAction()
}
