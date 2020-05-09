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
	"github.com/thediveo/go-asciitree"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
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
			// Look up the specified process for further use; bail out if it cannot be found.
			proc, ok := allns.Processes[pid]
			if !ok {
				return fmt.Errorf("unknown process PID %d", pid)
			}
			procPID = pid
			// Look up the specified target namespace, and bail out if we could
			// not discover it.
			tns, ok := allns.Namespaces[lxkns.TypeIndex(nst)][nsid]
			if !ok {
				return fmt.Errorf("unknown namespace %s", args[0])
			}
			// First determine whether the process will have no capabilities,
			// its effective capabilities, or even full capabilities. As a side
			// effect, this also gives us the process' effective UID, which
			// we'll later use when displaying the process node.
			tcaps, proceuid, err := caps(proc, tns)
			if err != nil {
				return err
			}
			// Next, create the separate process and target (node) branches,
			// then combine them to the extend possible for rendering.
			procbr, err := processbranch(proc, proceuid)
			if err != nil {
				return err
			}
			tbr := targetbranch(tns, tcaps)
			root := combine(procbr, tbr)
			// Finally, we can render this mess.
			fmt.Fprint(os.Stdout, // TODO: allow output redirection
				asciitree.Render(
					root,
					&NodeVisitor{},
					style.NamespaceStyler))
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
	rootCmd.PersistentFlags().BoolVar(&showProcCaps,
		"proccaps", true,
		"show the process' capabilities")
	rootCmd.PersistentFlags().BoolVar(&briefCaps,
		"brief", false,
		"show only a summary statement for the capabilities in the target namespace")
	cli.AddFlags(rootCmd)
	return
}

// Show or hide the process' capabilities; please don't confuse this with the
// capabilities of the process in the specified target namespace.
var showProcCaps bool

// Show either a capabilities list in the target, or just a short summary.
var briefCaps bool

var procPID lxkns.PIDType
