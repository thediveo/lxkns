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

package style

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"github.com/thediveo/lxkns"
)

// ProcessName returns the "name" of a process for display, based on the
// display mode in procNameMode.
func ProcessName(proc *lxkns.Process) string {
	switch procNameMode {
	case ProcName:
	case ProcBasename:
		return proc.Basename()
	case ProcExe:
		if len(proc.Cmdline) > 0 {
			return proc.Cmdline[0]
		}
	}
	return proc.Name
}

// The CLI flag controlling how to display process name.
var procNameMode = ProcName

// ProcNameMode is an enumeration setting the process name display mode.
type ProcNameMode enumflag.Flag

// Enumeration of allowed ProcNameMode values.
const (
	ProcName ProcNameMode = iota
	ProcBasename
	ProcExe
)

// Defines the textual representations for the ProcNameMode values.
var procNameModeIds = map[ProcNameMode][]string{
	ProcName:     {"name"},
	ProcBasename: {"basename"},
	ProcExe:      {"exe"},
}

func init() {
	// Delayed registration of our CLI flag.
	pflagCreators.Register(func(rootCmd *cobra.Command) {
		rootCmd.PersistentFlags().Var(
			enumflag.New(&procNameMode, "proc", procNameModeIds, enumflag.EnumCaseSensitive),
			"proc",
			"process name style; can be 'name' (default if omitted), 'basename',\n"+
				"or 'exe'")
		rootCmd.PersistentFlags().Lookup("proc").NoOptDefVal = "name"
	})
}