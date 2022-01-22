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

package namespaces

import (
	"fmt"
	"os"
	"os/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("namespaces", func() {

	Describe("plain namespaces", func() {

		It("render details", func() {
			pns := &PlainNamespace{
				nsid:   species.NamespaceID{Dev: 1, Ino: 123},
				nstype: species.CLONE_NEWNET,
			}
			Expect(pns.TypeIDString()).To(Equal("net:[123]"))
			Expect(pns.LeaderString()).To(Equal(""))

			pns.leaders = []*model.Process{
				{PID: 666, Starttime: 666, Name: "foo"},
				{PID: 42, Starttime: 42, Name: "bar"},
			}
			s := pns.String()
			Expect(s).To(ContainSubstring(`net:[123]`))
			Expect(s).To(ContainSubstring(`"foo" (666)`))
			Expect(s).To(ContainSubstring(`"bar" (42)`))

			pns.owner = &UserNamespace{
				HierarchicalNamespace: HierarchicalNamespace{
					PlainNamespace: PlainNamespace{
						nsid:   species.NamespaceID{Dev: 1, Ino: 777},
						nstype: species.CLONE_NEWUSER,
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
			pns := &PlainNamespace{
				leaders: []*model.Process{
					{PID: 666, Starttime: 666},
					{PID: 42, Starttime: 42},
				},
			}
			Expect(pns.Ealdorman()).To(Equal(&model.Process{PID: 42, Starttime: 42}))
		})

		It("find no ealdorman when there isn't one", func() {
			pns := &PlainNamespace{
				leaders: []*model.Process{},
			}
			Expect(pns.Ealdorman()).To(BeNil())
		})

		It("lives with errors when detecting the owner", func() {
			pns := &PlainNamespace{}
			Expect(func() { pns.DetectOwner(nil) }).NotTo(Panic())
		})

	})

	Describe("hierarchical namespaces", func() {

		It("render details", func() {
			hns := &HierarchicalNamespace{
				PlainNamespace: PlainNamespace{
					nsid:   species.NamespaceID{Dev: 1, Ino: 123},
					nstype: species.CLONE_NEWPID,
				},
			}
			s := hns.String()
			Expect(s).To(ContainSubstring("pid:[123]"))
			Expect(s).To(ContainSubstring("parent none"))
			Expect(s).To(ContainSubstring("children none"))

			chns := &HierarchicalNamespace{
				PlainNamespace: PlainNamespace{
					nsid:   species.NamespaceID{Dev: 1, Ino: 678},
					nstype: species.CLONE_NEWPID,
				},
			}
			hns.AddChild(chns)
			Expect(hns.String()).To(ContainSubstring("children [pid:[678]]"))
			Expect(chns.String()).To(ContainSubstring("parent pid:[123]"))
		})

		It("cant add a child namespace twice", func() {
			hns := &HierarchicalNamespace{}
			chns := &HierarchicalNamespace{}
			Expect(func() { hns.AddChild(chns) }).NotTo(Panic())
			Expect(func() { hns.AddChild(chns) }).To(Panic())
		})

	})

	Describe("user namespaces", func() {

		It("render details", func() {
			uns := New(species.CLONE_NEWUSER, species.NamespaceID{Dev: 1, Ino: 1111}, nil).(*UserNamespace)
			uns.owneruid = os.Getuid()
			uns.AddLeader(&model.Process{
				PID:  88888,
				Name: "foobar",
			})
			uns.ownedns[model.NetNS][species.NamespaceID{Dev: 1, Ino: 1234}] = &PlainNamespace{
				nsid:   species.NamespaceID{Dev: 1, Ino: 1234},
				nstype: species.CLONE_NEWNET,
			}

			Expect(uns.UID()).To(Equal(uns.owneruid))
			Expect(uns.Ownings()[model.NetNS]).To(HaveLen(1))

			s := uns.String()
			Expect(s).To(ContainSubstring(`joined by "foobar" (88888)`))
			Expect(s).To(ContainSubstring("owning [net:[1234]]"))
			Expect(s).To(ContainSubstring("user:[1111]"))
			u, _ := user.Current()
			Expect(s).To(ContainSubstring(
				fmt.Sprintf(`UID %d ("%s")`, os.Geteuid(), u.Username)))
		})

	})

	It("creates new namespace objects", func() {
		plainns := NewWithSimpleRef(species.CLONE_NEWNET, species.NamespaceID{Dev: 1, Ino: 1111}, "/foobar")
		Expect(plainns).To(BeAssignableToTypeOf(&PlainNamespace{}))
		Expect(plainns).NotTo(BeAssignableToTypeOf(&HierarchicalNamespace{}))
		Expect(plainns.Type()).To(Equal(species.CLONE_NEWNET))
		Expect(plainns.Ref()).To(ConsistOf("/foobar"))

		pidns := NewWithSimpleRef(species.CLONE_NEWPID, species.NamespaceID{Dev: 1, Ino: 1111}, "/foobar")
		Expect(pidns).To(BeAssignableToTypeOf(&HierarchicalNamespace{}))
		Expect(pidns).NotTo(BeAssignableToTypeOf(&UserNamespace{}))
		Expect(pidns.Type()).To(Equal(species.CLONE_NEWPID))
		Expect(pidns.Ref()).To(ConsistOf("/foobar"))
	})

})
