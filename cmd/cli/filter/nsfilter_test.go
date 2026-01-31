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

package filter

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/safe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("--filter flag", func() {

	It("rejects unknown namespaces", func() {
		cmd := cobra.Command{}
		cmd.SetArgs([]string{"--filter=foobar"})
		SetupCLI(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		Expect(cmd.Execute()).To(MatchError(MatchRegexp(`^invalid argument "foobar"`)))
	})

	It("gets namespace type list", func() {
		cmd := cobra.Command{}
		cmd.SetArgs([]string{"--filter=m,c,uts,U"})
		SetupCLI(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		Expect(cmd.Execute()).To(Succeed())

		flt := New(&cmd)
		matches := []species.NamespaceType{}
		for _, t := range defaultNamespaceFilters {
			if !flt(namespaces.NewWithSimpleRef(t, species.NamespaceID{}, "")) {
				continue
			}
			matches = append(matches, t)
		}
		Expect(matches).To(ConsistOf(
			species.CLONE_NEWNS,
			species.CLONE_NEWCGROUP,
			species.CLONE_NEWUTS,
			species.CLONE_NEWUSER,
		))
	})

})
