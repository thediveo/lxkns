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
	o "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// BeInAGroup succeeds if actual is a model.Container or *model.Container and
// the specified option matchers all succeed on a group the actual container
// belongs to.
//
//   Expect(c).To(BeInAGroup(WithName("my_project")))
func BeInAGroup(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withContainer("BeInAGroup",
		o.HaveField("Groups", o.ContainElement(o.SatisfyAll(opts...))))
}
