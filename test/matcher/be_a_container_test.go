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

var _ = Describe("BeAContainer matcher", func() {

	It("doesn't match something else", func() {
		Expect(BeADockerContainer().Match(nil)).Error().To(
			MatchError(ContainSubstring("expects a model.Container or *model.Container, but got <nil>")))
		Expect(BeAContainer().Match(42)).Error().To(
			MatchError(ContainSubstring("expects a model.Container or *model.Container, but got int")))
	})

	It("matches", func() {
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

		Expect(container).To(BeAContainer())
		Expect(container).To(BeAContainer(HaveField("ID", container.ID)))
	})

})
