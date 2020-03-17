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
	// Finally render the output based on the information gathered. The
	// important part here is the PIDVisitor, which encapsulated the knowledge
	// of traversing the information in the correct way in order to achieve
	// the desired process tree with PID namespaces.
	fmt.Println(
		asciitree.Render(
			allns.PIDNSRoots, &PIDVisitor{
				Details:      true,
				PIDMap:       pidmap,
				InitialPIDNS: allns.PIDNSRoots[0], //allns.InitialNamespaces[lxkns.PIDNS],
			}, common.NamespaceStyler))
}
