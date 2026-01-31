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
	"github.com/thediveo/enumflag/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("styles output", func() {

	DescribeTable("has the correct colormodes",
		func(arg string, expected string, colormode ColorMode, fail bool) {
			var cm ColorMode
			value := enumflag.New(&cm, "colormode", colorModeIds, enumflag.EnumCaseSensitive)
			err := value.Set(arg)
			if fail {
				Expect(err).To(HaveOccurred(), "should have failed")
				return
			}
			Expect(err).To(Succeed(), "should not have failed")
			Expect(value.String()).To(Equal(expected))
			Expect(cm).To(Equal(colormode))
		},
		Entry(nil, "foo", "", ColorMode(0), true),
		Entry(nil, "always", "always", ColorAlways, false),
		Entry(nil, "on", "always", ColorAlways, false),
		Entry(nil, "never", "never", ColorNever, false),
		Entry(nil, "off", "never", ColorNever, false),
		Entry(nil, "auto", "auto", ColorAuto, false),
	)

})
