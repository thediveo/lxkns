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
	"github.com/thediveo/enumflag/v2"
)

var _ = Describe("styles output", func() {

	It("has the correct colormodes", func() {
		tests := []struct {
			arg       string
			expected  string
			colormode ColorMode
			fail      bool
		}{
			{"foo", "", 0, true},
			{"always", "always", ColorAlways, false},
			{"on", "always", ColorAlways, false},
			{"never", "never", ColorNever, false},
			{"off", "never", ColorNever, false},
			{"auto", "auto", ColorAuto, false},
		}
		for _, tst := range tests {
			var cm ColorMode
			value := enumflag.New(&cm, "colormode", colorModeIds, enumflag.EnumCaseSensitive)
			err := value.Set(tst.arg)
			if tst.fail {
				Expect(err).To(HaveOccurred(), "should have failed")
				continue
			}
			Expect(err).To(Succeed(), "should not have failed")
			Expect(value.String()).To(Equal(tst.expected))
			Expect(cm).To(Equal(tst.colormode))
		}
	})

})
