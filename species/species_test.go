// Copyright 2020 Harald Albrecht.
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

package species

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace types", func() {

	It("stringify", func() {
		Expect(CLONE_NEWNS.String()).To(Equal("CLONE_NEWNS"))
		Expect(CLONE_NEWTIME.String()).To(Equal("CLONE_NEWTIME"))
		Expect((CLONE_NEWCGROUP | CLONE_NEWIPC).String()).
			To(Equal(fmt.Sprintf("NamespaceType(%d)", CLONE_NEWCGROUP|CLONE_NEWIPC)))
	})

	It("defines CLONE_NEWTIME", func() {
		Expect(CLONE_NEWTIME).To(Equal(NamespaceType(0x80)))
	})

})
