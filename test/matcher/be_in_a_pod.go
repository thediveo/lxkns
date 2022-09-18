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

// BeInAPod succeeds if actual is a model.Container or *model.Container and the
// container is grouped by a Kubernetes/k8s pod for which all the option
// matchers also succeed.
//
//	Expect(c).To(BeInAPod(WithName("default/mypod")))
func BeInAPod(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withContainer("BeInAPod",
		o.HaveField("Groups",
			o.ContainElement(BeAPod(opts...))))
}
