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

package cli

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli/cliplugin"
)

func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "cli_test",
		Group: cliplugin.Group,
		Symbols: []plugger.Symbol{
			plugger.NamedSymbol{Name: "SetupCLI", Symbol: UnittestSetupCLI},
			plugger.NamedSymbol{Name: "BeforeRun", Symbol: UnittestBeforeRun},
		},
	})
}

var setupCLI = 0

func UnittestSetupCLI(rootCmd *cobra.Command) {
	setupCLI++
}

var beforeRun = 0
var beforeRunErr = error(nil)

func UnittestBeforeRun() error {
	beforeRun++
	return beforeRunErr
}

var _ = Describe("CLI cmd plugins", func() {

	It("calls AddFlags plugin method", func() {
		rootCmd := cobra.Command{}
		rootCmd.SetArgs([]string{})
		setupCLI = 0
		AddFlags(&rootCmd)
		Expect(setupCLI).To(Equal(1))
	})

	It("calls BeforeCommand plugin method", func() {
		beforeRun = 0
		beforeRunErr = errors.New("fooerror")
		Expect(BeforeCommand()).To(HaveOccurred())
		Expect(beforeRun).To(Equal(1))

		beforeRunErr = nil
		Expect(BeforeCommand()).ToNot(HaveOccurred())
		Expect(beforeRun).To(Equal(2))
	})

})
