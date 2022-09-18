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

// WithType succeeds if actual has a Type field and optionally a Flavor field,
// and the specified typeflavor matches at least one of these fields. If you
// want to check only for the specific Type but not accept it as a Flavor, then
// use the WithStrictType matcher instead.
func WithType(typeflavor string) types.GomegaMatcher {
	return o.SatisfyAny(
		o.HaveField("Type", typeflavor),
		o.And(o.HaveExistingField("Flavor"), o.HaveField("Flavor", typeflavor)))
}
