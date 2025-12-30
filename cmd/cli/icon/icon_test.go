// Copyright 2023 Harald Albrecht.
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

package icon

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This unit test was writting during a unit test frenzy, triggered by zero
// trust. ha ha. HAHA. HAHAHAHAHA. ARGH!
var _ = Describe("icon CLI flag", Ordered, func() {

	It("defaults to not showing namespace type icons", func() {
		cmd := cobra.Command{}
		IconSetupCLI(&cmd)
		cmd.SetArgs([]string{})
		Expect(cmd.Execute()).To(Succeed())
		ni := NamespaceIcon(&cmd)
		Expect(ni(namespaces.NewWithSimpleRef(species.CLONE_NEWNS, species.NamespaceID{}, ""))).
			To(BeEmpty())
	})

	It("enables showing namespace type icons", func() {
		cmd := cobra.Command{}
		IconSetupCLI(&cmd)
		cmd.SetArgs([]string{"--" + IconFlagName})
		Expect(cmd.Execute()).To(Succeed())
		ni := NamespaceIcon(&cmd)
		Expect(ni(namespaces.NewWithSimpleRef(species.CLONE_NEWNS, species.NamespaceID{}, ""))).
			NotTo(BeEmpty())
	})

})
