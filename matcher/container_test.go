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

package matcher

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var container = model.Container{
	ID:     "1234567890",
	Name:   "foo_bar",
	Type:   "ducker.io",
	Flavor: "fluffy",
	Groups: []*model.Group{
		{Name: "fluffy", Type: "group.io", Flavor: "fluffy.io"},
	},
	Paused: true,
}

var _ = Describe("container", func() {

	It("doesn't match something else", func() {
		Expect(HaveName("42").Match(42)).Error().To(
			MatchError(ContainSubstring("expects a model.Container or *model.Container, but got int")))
	})

	It("matches container by name or ID", func() {
		Expect(container).NotTo(HaveName("rumpelpumpel"))

		Expect(container).To(HaveName(container.Name))
		Expect(container).To(HaveName(container.ID))

		Expect(&container).To(HaveName(container.Name))
	})

	It("matches container by name/ID and type/flavor", func() {
		Expect(container).NotTo(HaveNameAndType(container.Name, "rumpelpumpel"))

		Expect(container).To(HaveNameAndType(container.Name, container.Type))
		Expect(container).To(HaveNameAndType(container.Name, container.Flavor))

		Expect(&container).To(HaveNameAndType(container.Name, container.Type))
	})

	It("matches container by named group", func() {
		Expect(container).NotTo(HaveNamedGroup("iwo"))

		Expect(container).To(HaveNamedGroup(container.Groups[0].Name))
	})

	It("matches container by named group and type/flavor", func() {
		Expect(container).NotTo(HaveNamedAndTypedGroup(container.Groups[0].Name, "iwo"))

		Expect(container).To(HaveNamedAndTypedGroup(container.Groups[0].Name, container.Groups[0].Type))
		Expect(container).To(HaveNamedAndTypedGroup(container.Groups[0].Name, container.Groups[0].Flavor))
	})

	It("matches a paused container", func() {
		Expect(container).To(BePaused())

		c := container
		c.Paused = false
		Expect(c).NotTo(BePaused())
	})

})
