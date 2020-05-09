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
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/species"
)

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "nscaps [flags] NAMESPACE",
		Short:   "nscaps shows the capabilities of a process in a particular namespace",
		Version: lxkns.SemVersion,
		Args:    cobra.ExactArgs(1),
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.BeforeCommand()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nsid, nst := species.IDwithType(args[0])
			if nst == species.NaNS {
				return fmt.Errorf("not a valid namespace: %q", args[0])
			}
			fpid, _ := cmd.PersistentFlags().GetUint32("pid")
			pid := lxkns.PIDType(fpid)
			// Has a PID namespace different than our current one been
			// specified, in which the PID is valid?
			pidnsid := species.NoneID
			if nst, _ := cmd.PersistentFlags().GetString("ns"); nst != "" {
				id, err := strconv.ParseUint(nst, 10, 64)
				if err == nil {
					pidnsid, _ = species.IDwithType(fmt.Sprintf("pid:[%d]", id))
				} else {
					var t species.NamespaceType
					pidnsid, t = species.IDwithType(nst)
					if t != species.CLONE_NEWPID {
						return fmt.Errorf("not a valid PID namespace ID: %q", nst)
					}
				}
			}
			// Without a specific PID given, use our own PID as a substitution.
			if pid == 0 {
				if pidnsid != species.NoneID {
					return fmt.Errorf("--ns requires --pid=PID")
				}
				pid = lxkns.PIDType(os.Getpid())
			}
			// Run a full namespace discovery and also get the PID translation map.
			allns := lxkns.Discover(lxkns.FullDiscovery)
			pidmap := lxkns.NewPIDMap(allns)
			rootpidns := allns.Processes[lxkns.PIDType(os.Getpid())].Namespaces[lxkns.PIDNS]
			// If necessary, translate the PID from its own PID namespace into the
			// initial/this program's PID namespace.
			if pidnsid != species.NoneID {
				pidns := allns.Namespaces[lxkns.PIDNS][pidnsid]
				if pidns == nil {
					return fmt.Errorf("unknown PID namespace pid:[%d]", pidnsid.Ino)
				}
				rootpid := pidmap.Translate(lxkns.PIDType(pid), pidns, rootpidns)
				if rootpid == 0 {
					return fmt.Errorf("unknown process PID %d in pid:[%d]",
						pid, pidnsid.Ino)
				}
				pid = rootpid
			}
			//
			proc, ok := allns.Processes[pid]
			if !ok {
				return fmt.Errorf("unknown process PID %d", pid)
			}
			procbr, err := processbranch(proc)
			if err != nil {
				// FIXME:
				panic(err)
			}
			tns, ok := allns.Namespaces[lxkns.TypeIndex(nst)][nsid]
			if !ok {
				return fmt.Errorf("unknown namespace %s", args[0])
			}
			tbr := targetbranch(tns)
			root := combine(procbr, tbr)
			fmt.Printf("%+v", root)
			return nil
		},
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().Uint32P("pid", "p", 0,
		"PID of process for which to calculate capabilities")
	rootCmd.PersistentFlags().StringP("ns", "n", "",
		"PID namespace of PID, if not the initial PID namespace;\n"+
			"either an unsigned int64 value, such as \"4026531836\", or a\n"+
			"PID namespace textual representation like \"pid:[4026531836]\"")
	cli.AddFlags(rootCmd)
	return
}
