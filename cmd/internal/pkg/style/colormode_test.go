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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("styles output", func() {

	It("has a flag type", func() {
		var cm ColorMode
		Expect(cm.Type()).To(Equal("colormode"))
	})

	It("parses colormodes", func() {
		tests := []struct {
			arg       string
			expected  string
			colormode ColorMode
			fail      bool
		}{
			{"foo", "", 0, true},
			{"always", "always", CmAlways, false},
			{"on", "always", CmAlways, false},
			{"never", "never", CmNever, false},
			{"off", "never", CmNever, false},
			{"auto", "auto", CmAuto, false},
		}
		for _, tst := range tests {
			var cm ColorMode
			err := cm.Set(tst.arg)
			if tst.fail {
				Expect(err).To(HaveOccurred(), "should have failed")
				continue
			}
			Expect(err).NotTo(HaveOccurred(), "should not have failed")
			Expect(cm.String()).To(Equal(tst.expected))
			Expect(cm).To(Equal(tst.colormode))
		}
	})

})
