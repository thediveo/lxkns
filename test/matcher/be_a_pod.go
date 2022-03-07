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
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
)

// BeAPod succeeds if actual is a model.Group or a *model.Group and also
// satisfies all specified option matchers.
//
//   Expect(g).To(BeAPod(WithName("default/pod")))
//
// Related: the BeInAPod matchers checks a container to be part of a pod with
// specific properties.
func BeAPod(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withPod("BeAPod", o.SatisfyAll(
		o.HaveField("Type", kuhbernetes.PodGroupType),
		o.SatisfyAll(opts...)))
}
