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

import (
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns/model"
)

// ReexecIntoAction forks and then re-executes this process in order to run a
// specific action (indicated by actionname) in a set of (different) Linux
// kernel namespaces. The stdout result of running the action is then
// deserialized as JSON into the specified result element.
func ReexecIntoAction(actionname string, namespaces []model.Namespace, result interface{}) (err error) {
	return ReexecIntoActionEnv(actionname, namespaces, nil, result)
}

// ReexecIntoActionEnv forks and then re-executes this process in order to run
// a specific action (indicated by actionname) in a set of (different) Linux
// kernel namespaces. It also passes the additional environment variables
// specified in envvars. The stdout result of running the action is then
// deserialized as JSON into the specified result element.
func ReexecIntoActionEnv(actionname string, namespaces []model.Namespace, envvars []string, result interface{}) (err error) {
	rexns := make([]reexec.Namespace, len(namespaces))
	for idx := range namespaces {
		rexns[idx].Type = "!" + namespaces[idx].Type().Name()
		rexns[idx].Path = namespaces[idx].Ref()
	}
	return reexec.RunReexecAction(
		actionname,
		reexec.Namespaces(rexns),
		reexec.Environment(envvars),
		reexec.Result(result))
}
