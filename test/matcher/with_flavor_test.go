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
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

var _ = Describe("WithFlavor matcher", func() {

	It("matches container by flavor only", func() {
		var container = model.Container{
			Type:   moby.Type,
			Flavor: "moby",
		}

		Expect(container).To(WithFlavor(container.Flavor))
		Expect(container).NotTo(WithFlavor(container.Type))
	})

})
