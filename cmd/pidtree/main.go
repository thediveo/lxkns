// The "pidtree" CLI tool shows a simplified process tree, but with the
// following twists: it also shows PID namespaces, and translates PIDs into
// their PID namespace-local versions. In addition, it also allows showing
// only a single branch leading to a specific process/PID in a specific PID
// namespace.

// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"os"

	"github.com/thediveo/gons/reexec"
)

func main() {
	// For some discovery methods this app must be forked and re-executed; the
	// call to reexec.CheckAction() will automatically handle this situation
	// and then never return when in re-execution.
	reexec.CheckAction()
	// Otherwise, this is cobra boilerplate documentation, except for the
	// missing call to fmt.Println(err) which in the original boilerplate is
	// just plain wrong: it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		osExit(1)
	}
}

// For CLI unit tests...
var osExit = os.Exit
