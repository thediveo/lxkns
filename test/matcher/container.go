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
	wm "github.com/thediveo/whalewatcher/test/matcher"
)

// BeAContainer succeeds if actual is a model.Container or *model.Container and
// also satisfies all specified option matchers.
func BeAContainer(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withContainer("HaveContainer", o.SatisfyAll(opts...))
}

// WithName succeeds if actual has a Name field and optionally an ID field, and
// the specified nameid matches at least one of these fields.
func WithName(nameid string) types.GomegaMatcher {
	return o.SatisfyAny(o.HaveField("Name", nameid), wm.HaveOptionalField("ID", nameid))
}

// WithType succeeds if actual has Type and Flavor fields and the specified
// typeflavor matches at least one of these fields.
func WithType(typeflavor string) types.GomegaMatcher {
	return o.SatisfyAny(o.HaveField("Type", typeflavor), o.HaveField("Flavor", typeflavor))
}

// BeInGroup succeeds if actual has a Groups field and the specified options all
// succeed on one of the elements from the Groups field.
func BeInGroup(opts ...types.GomegaMatcher) types.GomegaMatcher {
	return withContainer("WithinGroup",
		o.HaveField("Groups", o.ContainElement(o.SatisfyAll(opts...))))
}
