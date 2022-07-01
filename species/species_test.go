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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace types", func() {

	DescribeTable("stringify",
		func(nstype NamespaceType, expected string) {
			Expect(nstype.String()).To(Equal(expected))
		},
		Entry("CLONE_NEWNS", CLONE_NEWNS, "CLONE_NEWNS"),
		Entry("CLONE_NEWTIME", CLONE_NEWTIME, "CLONE_NEWTIME"),
		Entry("CLONE_NEWCGROUP | CLONE_NEWIPC", CLONE_NEWCGROUP|CLONE_NEWIPC,
			fmt.Sprintf("NamespaceType(%d)", CLONE_NEWCGROUP|CLONE_NEWIPC)),
	)

	It("defines CLONE_NEWTIME", func() {
		Expect(CLONE_NEWTIME).To(Equal(NamespaceType(0x80)))
	})

})
