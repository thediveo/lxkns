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

package xstrings

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("spacing concatenated strings", func() {

	DescribeTable("spacing between string elements",
		func(x []string, expected string) {
			Expect(Join(x[0], x[1:]...)).To(Equal(expected))
		},
		Entry(nil, []string{"foo"}, "foo"),
		Entry(nil, []string{"foo", "bar", "baz"}, "foo bar baz"),
		Entry(nil, []string{"foo", "", "", "bar", ""}, "foo bar"),
	)

})
