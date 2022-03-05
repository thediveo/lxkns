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
	. "github.com/onsi/ginkgo/v2"
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
		Expect(BeAContainer().Match(42)).Error().To(
			MatchError(ContainSubstring("expects a model.Container or *model.Container, but got int")))
	})

	It("matches by name or ID", func() {
		Expect(container).NotTo(BeAContainer(WithName("rumpelpumpel")))

		Expect(container).To(BeAContainer(WithName(container.Name)))
		Expect(container).To(BeAContainer(WithName(container.ID)))

		Expect(&container).To(BeAContainer(WithName(container.Name)))
	})

	It("matches container by name/ID and type/flavor", func() {
		Expect(container).NotTo(BeAContainer(
			WithName(container.Name), WithType("rumpelpumpel")))

		Expect(container).To(BeAContainer(
			WithName(container.Name), WithType(container.Type)))
		Expect(container).To(BeAContainer(
			WithName(container.Name), WithType(container.Flavor)))

		Expect(&container).To(BeAContainer(
			WithName(container.Name), WithType(container.Type)))
	})

	It("matches container by named group", func() {
		Expect(container).NotTo(BeInGroup(WithName("iwo")))

		Expect(container).To(BeInGroup(WithName(container.Groups[0].Name)))
	})

	It("matches container by named group and type/flavor", func() {
		Expect(container).NotTo(BeInGroup(
			WithName(container.Groups[0].Name), WithType("iwo")))

		Expect(container).To(BeInGroup(
			WithName(container.Groups[0].Name), WithType(container.Groups[0].Type)))
		Expect(container).To(BeInGroup(
			WithName(container.Groups[0].Name), WithType(container.Groups[0].Flavor)))
	})

})
