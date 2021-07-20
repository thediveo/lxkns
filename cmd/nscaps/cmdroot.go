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
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-asciitree"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/pkg/engines"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "nscaps [flags] NAMESPACE",
		Short:   "nscaps shows the capabilities of a process in a particular namespace",
		Version: lxkns.SemVersion,
		Args: func(cmd *cobra.Command, args []string) error {
			if dump, _ := cmd.PersistentFlags().GetBool("dump"); dump {
				// --dump ignores any arguments, as it dumps a theme to stdout
				// and then exits.
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("expects 1 arg, received %d", len(args))
			}
			return nil
		},
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.BeforeCommand()
		},
		RunE: nscapscmd,
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

// PID of process whose capabilities in some namespace are sought.
var procPID model.PIDType

// Determine the capabilities of a process (either this one or another
// explicitly specified one) in another namespace. It then renders a nice tree
// consisting of the branches with the process and namespace, as well as
// indications of accessibility depending the where the process and a particular
// namespace are.
func nscapscmd(cmd *cobra.Command, args []string) error {
	nsid, nst := species.IDwithType(args[0])
	if nst == species.NaNS {
		return fmt.Errorf("not a valid namespace: %q", args[0])
	}
	fpid, _ := cmd.PersistentFlags().GetUint32("pid")
	pid := model.PIDType(fpid)
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
				return fmt.Errorf("not a valid PID namespace: %q", nst)
			}
		}
	}
	// Without a specific PID given, use our own PID as a substitution.
	if pid == 0 {
		if pidnsid != species.NoneID {
			return fmt.Errorf("--ns requires --pid=PID")
		}
		pid = model.PIDType(os.Getpid())
	}
	// Run a full namespace discovery and also get the PID translation map.
	cizer, err := engines.Containerizer(context.Background(), cmd, true)
	if err != nil {
		return err
	}
	allns := lxkns.Discover(lxkns.WithStandardDiscovery(), lxkns.WithContainerizer(cizer))
	pidmap := lxkns.NewPIDMap(allns)
	rootpidns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.PIDNS]
	// If necessary, translate the PID from its own PID namespace into the
	// initial/this program's PID namespace.
	if pidnsid != species.NoneID {
		pidns := allns.Namespaces[model.PIDNS][pidnsid]
		if pidns == nil {
			return fmt.Errorf("unknown PID namespace pid:[%d]", pidnsid.Ino)
		}
		rootpid := pidmap.Translate(model.PIDType(pid), pidns, rootpidns)
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
	tns, ok := allns.Namespaces[model.TypeIndex(nst)][nsid]
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
	fmt.Fprint(os.Stdout,
		asciitree.Render(
			root,
			&NodeVisitor{},
			style.NamespaceStyler))
	return nil
}
