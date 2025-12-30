// Copyright 2023, 2025 Harald Albrecht.
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

package xslices

import (
	"strings"

	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sorting", func() {

	DescribeTable("sorting a copy",
		func(x []string, expected []string) {
			original := slices.Clone(x)
			actual := SortedCopy(x, strings.Compare)
			Expect(actual).To(ConsistOf(expected), "not correctly sorted")
			Expect(x).To(ConsistOf(original), "modified original")
		},
		Entry(nil, []string{"bar", "foo"}, []string{"bar", "foo"}),
		Entry(nil, []string{"foo", "bar"}, []string{"bar", "foo"}),
	)

})
