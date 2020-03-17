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
	"fmt"
	"os"
	"os/user"

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

		It("render details", func() {
			pns := &plainNamespace{
				nsid:   123,
				nstype: nstypes.CLONE_NEWNET,
			}
			Expect(pns.TypeIDString()).To(Equal("net:[123]"))
			Expect(pns.LeaderString()).To(Equal(""))

			pns.leaders = []*Process{
				{PID: 666, Starttime: 666, Name: "foo"},
				{PID: 42, Starttime: 42, Name: "bar"},
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

		It("find an ealdorman", func() {
			pns := &plainNamespace{
				leaders: []*Process{
					{PID: 666, Starttime: 666},
					{PID: 42, Starttime: 42},
				},
			}
			Expect(pns.Ealdorman()).To(Equal(&Process{PID: 42, Starttime: 42}))
		})

		It("find no ealdorman when there isn't one", func() {
			pns := &plainNamespace{
				leaders: []*Process{},
			}
			Expect(pns.Ealdorman()).To(BeNil())
		})

		It("lives with errors when detecting the owner", func() {
			pns := &plainNamespace{}
			Expect(func() { pns.DetectOwner(nil) }).NotTo(Panic())
		})

	})

	Describe("hierarchical namespaces", func() {

		It("render details", func() {
			hns := &hierarchicalNamespace{
				plainNamespace: plainNamespace{
					nsid:   123,
					nstype: nstypes.CLONE_NEWPID,
				},
			}
			s := hns.String()
			Expect(s).To(ContainSubstring("pid:[123]"))
			Expect(s).To(ContainSubstring("parent none"))
			Expect(s).To(ContainSubstring("children none"))

			chns := &hierarchicalNamespace{
				plainNamespace: plainNamespace{
					nsid:   678,
					nstype: nstypes.CLONE_NEWPID,
				},
			}
			hns.AddChild(chns)
			Expect(hns.String()).To(ContainSubstring("children [pid:[678]]"))
			Expect(chns.String()).To(ContainSubstring("parent pid:[123]"))
		})

		It("cant add a child namespace twice", func() {
			hns := &hierarchicalNamespace{}
			chns := &hierarchicalNamespace{}
			Expect(func() { hns.AddChild(chns) }).NotTo(Panic())
			Expect(func() { hns.AddChild(chns) }).To(Panic())
		})

	})

	Describe("user namespaces", func() {

		It("render details", func() {
			uns := NewNamespace(nstypes.CLONE_NEWUSER, nstypes.NamespaceID(1111), "").(*userNamespace)
			uns.owneruid = os.Getuid()
			uns.AddLeader(&Process{
				PID:  88888,
				Name: "foobar",
			})
			uns.ownedns[NetNS][1234] = &plainNamespace{
				nsid:   1234,
				nstype: nstypes.CLONE_NEWNET,
			}

			Expect(uns.UID()).To(Equal(uns.owneruid))
			Expect(uns.Ownings()[NetNS]).To(HaveLen(1))

			s := uns.String()
			Expect(s).To(ContainSubstring(`joined by "foobar" (88888)`))
			Expect(s).To(ContainSubstring("owning [net:[1234]]"))
			Expect(s).To(ContainSubstring("user:[1111]"))
			u, _ := user.Current()
			Expect(s).To(ContainSubstring(
				fmt.Sprintf(`UID %d ("%s")`, os.Geteuid(), u.Username)))
		})

	})

})
