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

package lxkns

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstypes"
)

var _ = Describe("namespaces", func() {

	It("TypeIndex fails for invalid kernel namespace type", func() {
		Expect(TypeIndex(nstypes.CLONE_NEWCGROUP | nstypes.CLONE_NEWNET)).To(
			Equal(NamespaceTypeIndex(-1)))
	})

	Describe("plain namespaces", func() {

		It("describe themselves", func() {
			pns := &plainNamespace{
				nsid:   123,
				nstype: nstypes.CLONE_NEWNET,
			}
			Expect(pns.TypeIDString()).To(Equal("net:[123]"))

			pns.leaders = []*Process{
				&Process{PID: 666, Starttime: 666, Name: "foo"},
				&Process{PID: 42, Starttime: 42, Name: "bar"},
			}
			s := pns.String()
			Expect(s).To(ContainSubstring(`net:[123]`))
			Expect(s).To(ContainSubstring(`"foo" (666)`))
			Expect(s).To(ContainSubstring(`"bar" (42)`))

			pns.owner = &userNamespace{
				hierarchicalNamespace: hierarchicalNamespace{
					plainNamespace: plainNamespace{
						nsid:   777,
						nstype: nstypes.CLONE_NEWUSER,
					},
				},
			}
			s = pns.String()
			Expect(s).To(ContainSubstring(`net:[123]`))
			Expect(s).To(ContainSubstring(`"foo" (666)`))
			Expect(s).To(ContainSubstring(`"bar" (42)`))
			Expect(s).To(ContainSubstring(`user:[777]`))
		})

		It("finds the ealdorman", func() {
			pns := &plainNamespace{
				leaders: []*Process{
					&Process{PID: 666, Starttime: 666},
					&Process{PID: 42, Starttime: 42},
				},
			}
			Expect(pns.Ealdorman()).To(Equal(&Process{PID: 42, Starttime: 42}))
		})

		It("finds no ealdorman when there isn't one", func() {
			pns := &plainNamespace{
				leaders: []*Process{},
			}
			Expect(pns.Ealdorman()).To(BeNil())
		})

	})

})
