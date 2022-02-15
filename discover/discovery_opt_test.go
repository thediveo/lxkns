// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux
// +build linux

package discover

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func withOptions(options ...DiscoveryOption) DiscoverOpts {
	opts := DiscoverOpts{
		Labels: map[string]string{},
	}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

var _ = Describe("discovery options", func() {

	It("adds label(s)", func() {
		Expect(withOptions().Labels).To(BeEmpty())
		Expect(withOptions(
			WithLabel("foo", "bar"),
			WithLabels(map[string]string{"bar": "baz"}),
		).Labels).To(And(HaveKeyWithValue("foo", "bar"), HaveKeyWithValue("bar", "baz")))
	})

})
