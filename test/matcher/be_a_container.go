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
	o "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// BeAContainer succeeds if actual is a model.Container or *model.Container and
// also satisfies all specified option matchers. A typical use is to use
// BeAContainer in combination with the "WithX" matchers, such as WithName and
// WithType. Of course, any other matcher can be specified as an option matcher
// to BeAContainer as needed.
//
//   Expect(c).To(BeAContainer(WithName("foobar"), WithType(moby.Type),
//       BePaused()))
//
// Checking for the presence of certain containers in slices, arrays and maps is
// straightforward: simply list the expected combination of container
// properties, such as not only a specific name, but also being part of a
// particular Kubernetes pod or Composer project.
//
//   Expect(containers).To(ContainElement(BeAContainer(
//       WithName("foobar"), BeInAPod(WithName("default/pod")))))
func BeAContainer(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withContainer("HaveContainer", o.SatisfyAll(opts...))
}
