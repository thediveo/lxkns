// The "pidtree" CLI tool shows a simplified process tree, but with the
// following twists: it also shows PID namespaces, and translates PIDs into
// their PID namespace-local versions.

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

	asciitree "github.com/thediveo/go-asciitree"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns"
	common "github.com/thediveo/lxkns/cmd/internal/pkg/shared"
)

func main() {
	// For some discovery methods this app must be forked and re-executed; the
	// call to reexec.CheckAction() will automatically handle this situation
	// and then never return when in re-execution.
	reexec.CheckAction()
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
	rootpidns := allns.Processes[lxkns.PIDType(os.Getpid())].Namespaces[lxkns.PIDNS]
	// Finally render the output based on the information gathered. The
	// important part here is the PIDVisitor, which encapsulated the knowledge
	// of traversing the information in the correct way in order to achieve
	// the desired process tree with PID namespaces.
	fmt.Println(
		asciitree.Render(
			[]lxkns.Namespace{rootpidns}, // note to self: expects a slice of roots
			&PIDVisitor{
				Details:   true,
				PIDMap:    pidmap,
				RootPIDNS: rootpidns,
			},
			common.NamespaceStyler))
}
