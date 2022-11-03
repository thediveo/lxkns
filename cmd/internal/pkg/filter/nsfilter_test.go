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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("--filter flag", func() {

	It("rejects unknown namespaces", func() {
		rootCmd := cobra.Command{}
		rootCmd.SetArgs([]string{"--filter=foobar"})
		SetupCLI(&rootCmd)
		var err error
		_ = getstdout.Stdouterr(func() { err = rootCmd.Execute() })
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(`^invalid argument "foobar"`))
	})

	It("gets namespace type list", func() {
		rootCmd := cobra.Command{}
		rootCmd.SetArgs([]string{"--filter=m,c,uts,U"})
		SetupCLI(&rootCmd)
		Expect(rootCmd.Execute()).ToNot(HaveOccurred())
		Expect(namespaceFilters).To(HaveLen(4))
		Expect(namespaceFilters).To(ContainElements(
			species.CLONE_NEWNS,
			species.CLONE_NEWCGROUP,
			species.CLONE_NEWUTS,
			species.CLONE_NEWUSER,
		))
		Expect(Filter(namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NamespaceID{}, ""))).To(BeFalse())
		Expect(Filter(namespaces.NewWithSimpleRef(species.CLONE_NEWUSER, species.NamespaceID{}, ""))).To(BeTrue())
	})

})
