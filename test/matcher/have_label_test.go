// Copyright 2022 Harald Albrecht.
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

package matcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("label matcher", func() {

	var container = model.Container{
		ID:     "1234567890",
		Name:   "foo_bar",
		Type:   "ducker.io",
		Flavor: "fluffy",
		Labels: model.Labels{
			"foo": "FOO",
		},
	}

	It("matches labels", func() {
		Expect(container).To(HaveLabel("foo"))
		Expect(container).To(HaveLabel("foo", "FOO"))
		Expect(container).NotTo(HaveLabel("bar"))
		Expect(container).NotTo(HaveLabel("foo", "BAR"))
	})

	It("panics when given multiple values", func() {
		Expect(func() { HaveLabel("foo", "bar", "baz") }).To(PanicWith(ContainSubstring("HaveLabel accepts")))
	})

})
