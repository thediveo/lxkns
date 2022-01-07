// Copyright 2021 Harald Albrecht.
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

package plural

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace types", func() {

	It("returns correct english form", func() {
		testdata := []struct {
			count    int
			elements string
			a        interface{}
			expected string
		}{
			{0, "containers", nil, "0 containers"},
			{1, "containers", nil, "1 container"},
			{42, "containers", nil, "42 containers"},
			{0, "pods", nil, "0 pods"},
			{1, "pods", nil, "1 pod"},
			{666, "pods", nil, "666 pods"},
			{1, "hidden %s namespaces", "foobar", "1 hidden foobar namespace"},
			{2, "hidden %s namespaces", "foobar", "2 hidden foobar namespaces"},
			{0, "%s namespaces", "mount", "0 mount namespaces"},
			{1, "%s namespaces", "mount", "1 mount namespace"},
		}

		for _, td := range testdata {
			if td.a != nil {
				Expect(Elements(td.count, td.elements, td.a)).To(Equal(td.expected))
			} else {
				Expect(Elements(td.count, td.elements)).To(Equal(td.expected))
			}
		}
	})

})
