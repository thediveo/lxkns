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

var _ = Describe("WithName matcher", func() {

	It("matches container by name or ID", func() {
		var container = model.Container{
			ID:   "1234567890",
			Name: "foo_bar",
		}

		Expect(container).To(WithName(container.Name))
		Expect(container).To(WithName(container.ID))

		Expect(container).NotTo(WithName("DOH!"))
	})

	It("matches group by name", func() {
		var group = model.Group{
			Name: "group1",
		}

		Expect(group).To(WithName(group.Name))

		Expect(group).NotTo(WithName("DOH!"))
	})

})
