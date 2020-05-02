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
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	asciitree "github.com/thediveo/go-asciitree"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/species"
)

// We only have the root command, but no (sub) commands, as pidtree is a
// simple command and not trying to become "ps".
var rootCmd = &cobra.Command{
	Use:   "pidtree",
	Short: "pidtree shows the tree of PID namespaces together with PIDs",
	Args:  cobra.NoArgs,
	Example: `  pidtree
	shows the PID namespaces hierarchy with the process inside them as a tree.
  pidtree -p 42
	shows only those PID namespaces hierarchy and processes on the branch
	leading to process PID 42.
  pidtree -n pid:[4026531836] -p 1
	shows only the PID namespace hierarchy and processes on the branch
	leading to process PID 1 in PID namespace 4026531836.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return cli.BeforeCommand()
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		pid, _ := cmd.PersistentFlags().GetUint32("pid")
		// If no PID was specified ("zero" PID), then render the usual full
		// PID namespace and process tree.
		if pid == 0 {
			return renderPIDTreeWithNamespaces(os.Stdout)
		}
		// If there is a PID, then check next if there is also a PID namespace
		// specified, in which the PID is valid. Then render only the branch
		// leading from the initial PID namespace down to the PID namespace of
		// PID, and the processes on this branch.
		sloppypidnsid := species.NoneID
		if nst, _ := cmd.PersistentFlags().GetString("ns"); nst != "" {
			id, err := strconv.ParseUint(nst, 10, 64)
			if err == nil {
				sloppypidnsid = species.NamespaceID{Ino: id}
			} else {
				var t species.NamespaceType
				sloppypidnsid, t = species.IDwithType(nst)
				if t == species.NaNS {
					return fmt.Errorf("not a valid PID namespace ID: %q", nst)
				}
			}
		}
		return renderPIDBranch(os.Stdout, lxkns.PIDType(pid), sloppypidnsid)
	},
}

// Sets up the flags.
func init() {
	rootCmd.PersistentFlags().Uint32P("pid", "p", 0,
		"PID of process to show PID namespace tree and parent PIDs for")
	rootCmd.PersistentFlags().StringP("ns", "n", "",
		"PID namespace of PID, if not the initial PID namespace;\n"+
			"either an unsigned int64 value, such as \"4026531836\", or a\n"+
			"PID namespace textual representation like \"pid:[4026531836]\"")
	cli.AddFlags(rootCmd)
}

// SingleBranch encodes a single branch from the initial/root PID namespace
// down to a particular process, with all intermediate PID namespaces and
// processes along the route.
type SingleBranch struct {
	Branch []interface{}
}

// Renders only the PID namespaces hierarchy and PID branch leading up to a
// specific PID, optionally in a specific PID namespace.
func renderPIDBranch(out io.Writer, pid lxkns.PIDType, sloppypidnsid species.NamespaceID) error {
	// Run a full namespace discovery and also get the PID translation map.
	allns := lxkns.Discover(lxkns.FullDiscovery)
	pidmap := lxkns.NewPIDMap(allns)
	rootpidns := allns.Processes[lxkns.PIDType(os.Getpid())].Namespaces[lxkns.PIDNS]
	// If necessary, translate the PID from its own PID namespace into the
	// initial/this program's PID namespace.
	if sloppypidnsid != species.NoneID {
		pidns := allns.Namespaces[lxkns.PIDNS].SloppyByIno(sloppypidnsid)
		if pidns == nil {
			return fmt.Errorf("unknown PID namespace pid:[%d]", sloppypidnsid.Ino)
		}
		rootpid := pidmap.Translate(pid, pidns, rootpidns)
		if rootpid == 0 {
			return fmt.Errorf("unknown process PID %d in pid:[%d]",
				pid, sloppypidnsid.Ino)
		}
		pid = rootpid
	}
	// Find the process with PID and then create just the render branch
	// leading to and terminating at it.
	proc, ok := allns.Processes[pid]
	if !ok {
		return fmt.Errorf("unknown process PID %d", pid)
	}
	branch := SingleBranch{Branch: []interface{}{}}
	for proc != nil {
		// Prepend the current process to the branch.
		branch.Branch = append([]interface{}{proc}, branch.Branch...)
		// Now if there is a change in PID namespaces just at the current
		// process, prepend our "current" PID namespace also. The difficult
		// part here is that we need to deal with the situation where we have
		// the process tree, but lack the PID namespace information for
		// processes in the tree and up the branch for which we don't have
		// enough privileges: we then cannot give PID namespace information
		// for them :(
		pproc := proc.Parent
		if (pproc == nil ||
			pproc.Namespaces[lxkns.PIDNS] != proc.Namespaces[lxkns.PIDNS]) &&
			proc.Namespaces[lxkns.PIDNS] != nil {
			branch.Branch = append(
				[]interface{}{proc.Namespaces[lxkns.PIDNS]},
				branch.Branch...)
		}
		// Climb up towards the root/stem.
		proc = pproc
	}
	// Now render the whole branch...
	fmt.Fprintln(out,
		asciitree.Render(
			branch,
			&BranchVisitor{
				Details:   true,
				PIDMap:    pidmap,
				RootPIDNS: rootpidns,
			},
			style.NamespaceStyler))
	return nil
}

// Renders a full PID tree including PID namespaces.
func renderPIDTreeWithNamespaces(out io.Writer) error {
	// Run a full namespace discovery and also get the PID translation map.
	allns := lxkns.Discover(lxkns.FullDiscovery)
	pidmap := lxkns.NewPIDMap(allns)
	// You may wonder why lxkns returns a slice of "root" PID and user
	// namespaces, instead of only a single root for each. The rationale is
	// that in some situation without sufficient privileges (capabilities) and
	// bind-mounted or fd-references PID and/or user namespaces, these can
	// still show up in the discovery process. We don't filter them out on
	// purpose. However, we might not be able to correlate them to processes,
	// as insufficient privileges (missing CAP_SYS_PTRACE) hinders us to read
	// the namespaces a process of another user is attached to. In
	// consequence, here we only start with our own PID namespace, ignoring
	// any other roots that might have turned up during discovery. And this
	// slightly ranty comment now gets me another badge-achievement which is
	// so important in today's societies: "ranty source commenter".
	ourproc, ok := allns.Processes[lxkns.PIDType(os.Getpid())]
	if !ok {
		fmt.Fprintln(os.Stderr, "error: /proc does not match the current PID namespace")
		os.Exit(1)
	}
	rootpidns := ourproc.Namespaces[lxkns.PIDNS]
	// Finally render the output based on the information gathered. The
	// important part here is the PIDVisitor, which encapsulated the knowledge
	// of traversing the information in the correct way in order to achieve
	// the desired process tree with PID namespaces.
	fmt.Fprintln(out,
		asciitree.Render(
			[]lxkns.Namespace{rootpidns}, // note to self: expects a slice of roots
			&TreeVisitor{
				Details:   true,
				PIDMap:    pidmap,
				RootPIDNS: rootpidns,
			},
			style.NamespaceStyler))
	return nil
}
