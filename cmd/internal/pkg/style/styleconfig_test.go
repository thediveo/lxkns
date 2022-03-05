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

package style

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
)

var _ = Describe("reads styles", func() {

	oldStyles := map[string]Style{}

	BeforeEach(func() {
		for sty, style := range Styles {
			oldStyles[sty] = *style
			*(Styles[sty]) = Style{}
		}
	})

	AfterEach(func() {
		for sty, style := range oldStyles {
			*(Styles[sty]) = style
		}
	})

	It("warns about invalid YAML", func() {
		Expect(getstdout.Stdouterr(func() { parseStyles("/*invalid*/") })).To(
			MatchRegexp(`^error: failed to parse style configuration yaml`))
	})

	It("warns about unknown styles", func() {
		Expect(getstdout.Stdouterr(func() { parseStyles("badstyle: foobar") })).To(
			MatchRegexp(`^warning: unknown element "badstyle"`))
	})

	It("handles style attributes", func() {
		Expect(getstdout.Stdouterr(func() {
			parseStyles(`
pid:
  - bold
  - underline
  - spanishinquisition
`)
		})).To(MatchRegexp(`^warning: unknown styling attribute "spanishinquisition"`))
		Expect(PIDStyle.S("x")).To(ContainSubstring("[1;4mx"))
	})

	It("rejects unknown colors", func() {
		Expect(getstdout.Stdouterr(func() {
			parseStyles(`
pid:
  - spanishinquisition: "#666"
`)
		})).To(MatchRegexp(`^warning: unknown color type "spanishinquisition"`))
	})

	It("handles colors", func() {
		Expect(getstdout.Stdouterr(func() {
			parseStyles(`
pid:
  - foreground: "#123123"
  - background: "#deadbf"
  - spanishinquisition: 666
`)
		})).To(MatchRegexp(`^warning: unknown color spanishinquisition: 666`))
		Expect(len(PIDStyle.S("x"))).To(BeNumerically(">", 1))
		Expect(PIDStyle.S("x")).To(ContainSubstring("mx"))
	})

})
