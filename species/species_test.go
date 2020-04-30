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

var _ = Describe("Namespace Types and IDs", func() {

	It("parse namespace textual representations", func() {
		id, t := IDwithType("net:[1]")
		Expect(t).To(Equal(CLONE_NEWNET))
		Expect(id).To(Equal(NamespaceID(1)))
	})

	It("reject invalid textual representations", func() {
		for _, text := range []string{
			"foo:[1]", "net:[-1]", "net[1]", "n:[1]", "net:[1",
		} {
			id, t := IDwithType(text)
			Expect(t).To(Equal(NaNS), "%s is not a namespace", text)
			Expect(id).To(Equal(NoneID), "%s is not a namespace", text)
		}
	})

	It("stringify", func() {
		Expect(CLONE_NEWNS.String()).To(Equal("CLONE_NEWNS"))
		Expect((CLONE_NEWCGROUP | CLONE_NEWIPC).String()).
			To(Equal(fmt.Sprintf("NamespaceType(%d)", CLONE_NEWCGROUP|CLONE_NEWIPC)))
		Expect(NamespaceID(123).String()).To(Equal("NamespaceID(123)"))
	})

})
