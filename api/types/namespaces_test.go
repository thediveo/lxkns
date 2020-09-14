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

package types

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstest/gmodel"
	. "github.com/thediveo/lxkns/nstest/gmodel"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var (
	allns      *lxkns.DiscoveryResult
	scripts    = testbasher.Basher{}
	scriptscmd *testbasher.TestCommand
	userns     model.Namespace
)

var _ = BeforeSuite(func() {
	scripts.Common(nstest.NamespaceUtilsScript)
	scripts.Script("main", `
unshare -Ur unshare -U $stage2 # create a new user ns inside another user ns (so we get a proper owner relationship).
	`)
	scripts.Script("stage2", `
process_namespaceid user # prints the user namespace ID of "the" process.
read # wait for test to proceed()
`)
	scriptscmd = scripts.Start("main")
	var usernsid species.NamespaceID
	scriptscmd.Decode(&usernsid)

	disco := lxkns.FullDiscovery
	disco.SkipBindmounts = true
	disco.SkipFds = true
	disco.SkipTasks = true
	allns = lxkns.Discover(disco) // "nearlyallns"

	userns = allns.Namespaces[model.UserNS][usernsid].(model.Namespace)
	Expect(userns).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	if scriptscmd != nil {
		scriptscmd.Close()
	}
	scripts.Done()
})

func pidlist(pids []model.PIDType) string {
	s := []string{}
	for _, pid := range pids {
		s = append(s, fmt.Sprint(pid))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(s, ", "))
}

func childlist(hns model.Hierarchy) string {
	s := []string{}
	for _, child := range hns.Children() {
		s = append(s, fmt.Sprint(child.(model.Namespace).ID().Ino))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(s, ", "))
}

var _ = Describe("namespaces JSON", func() {

	It("always gets a Namespace from the dictionary", func() {
		d := NewNamespacesDict()
		uns := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(123), "/foobar")
		d.AllNamespaces[model.UserNS][uns.ID()] = uns

		ns := d.Get(uns.ID(), uns.Type())
		Expect(ns).To(BeIdenticalTo(uns))

		ns = d.Get(species.NamespaceIDfromInode(666), species.CLONE_NEWNET)
		Expect(ns).NotTo(BeNil())
		Expect(ns.ID()).To(Equal(species.NamespaceIDfromInode(666)))
		Expect(ns.Type()).To(Equal(species.CLONE_NEWNET))
		Expect(ns.Ref()).To(BeZero())
	})

	It("marshals Namespace", func() {
		// A non-user and non-PID namespace must not contain parent and UID
		// information. But it must contain a owner user namespace reference.
		ns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS]
		Expect(ns).NotTo(BeNil())
		d := NewNamespacesDict()
		j, err := d.MarshalNamespace(ns)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "net",
				"owner": %d,
				"reference": %q,
				"leaders": %s
			}`,
			ns.ID().Ino,
			ns.Owner().(model.Namespace).ID().Ino,
			ns.Ref(),
			pidlist(ns.LeaderPIDs()),
		)))

		// In contrast, a user namespace must contain parent and UID
		// information. But it must not contain an owner reference, this is
		// the parent reference instead.
		parentuserns := userns.(model.Hierarchy).Parent().(model.Namespace)
		j, err = d.MarshalNamespace(userns)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				"reference": %q,
				"leaders": %s,
				"parent": %d,
				"user-uid": %d
			}`,
			userns.ID().Ino,
			userns.Ref(),
			pidlist(userns.LeaderPIDs()),
			parentuserns.ID().Ino,
			userns.(model.Ownership).UID(),
		)))

		// Check for the correct child list of the parent user namespace.
		j, err = d.MarshalNamespace(parentuserns)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				"reference": %q,
				"parent": %d,
				"children": %s,
				"user-uid": %d
			}`,
			parentuserns.ID().Ino,
			parentuserns.Ref(),
			parentuserns.(model.Hierarchy).Parent().(model.Namespace).ID().Ino,
			childlist(parentuserns.(model.Hierarchy)),
			parentuserns.(model.Ownership).UID(),
		)))

		// Also check the grandparent user namespace.
		grandpa := parentuserns.(model.Hierarchy).Parent().(model.Namespace)
		j, err = d.MarshalNamespace(grandpa)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				"reference": %q,
				"leaders": %s,
				"children": %s,
				"user-uid": %d
			}`,
			grandpa.ID().Ino,
			grandpa.Ref(),
			pidlist(grandpa.LeaderPIDs()),
			childlist(grandpa.(model.Hierarchy)),
			grandpa.(model.Ownership).UID(),
		)))
	})

	It("unmarshals Namespace", func() {
		d := NewNamespacesDict()
		// This unmarshalling MUST fail...
		_, err := d.UnmarshalNamespace([]byte(`""`))
		Expect(err).To(HaveOccurred())
		_, err = d.UnmarshalNamespace([]byte(`{"type":"foobar"}`))
		Expect(err).To(HaveOccurred())
		_, err = d.UnmarshalNamespace([]byte(`{"nsid":0,"type":"net"}`))
		Expect(err).To(HaveOccurred())

		// First create a JSON textual representation for a user namespace we
		// want to unmarshal next...
		j, err := d.MarshalNamespace(userns)
		Expect(err).To(Succeed())

		// ...now check that unmarshalling correctly works.
		nsdict := NewNamespacesDict()
		uns, err := nsdict.UnmarshalNamespace(j)
		Expect(err).To(Succeed())
		Expect(uns).To(BeSameNamespace(userns))

		// Check that unmarshalling a (flat) namespace also works correctly.
		ns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS]
		j, err = nsdict.MarshalNamespace(ns)
		Expect(err).To(Succeed())

		ns2, err := nsdict.UnmarshalNamespace(j)
		Expect(err).To(Succeed())
		Expect(ns2).To(gmodel.BeSameNamespace(ns))
	})

	It("marshals NamespacesDict", func() {
		d := NewNamespacesDict()
		d.AllNamespaces[model.UserNS][userns.ID()] = userns
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
			"%d": {
				"nsid": %[1]d,
				"type": "user",
				"reference": %q,
				"leaders": %s,
				"parent": %d,
				"user-uid": %d
			}
		}`,
			userns.ID().Ino,
			userns.Ref(),
			pidlist(userns.LeaderPIDs()),
			userns.(model.Hierarchy).Parent().(model.Namespace).ID().Ino,
			userns.(model.Ownership).UID(),
		)))
	})

	It("unmarshals NamespacesDict", func() {
		d := NewNamespacesDict()
		// This must NOT succeed...
		Expect(d.UnmarshalJSON([]byte(`""`))).To(HaveOccurred())
		Expect(d.UnmarshalJSON([]byte(`{"123":{"type":"foobar"}}`))).To(HaveOccurred())

		// To unmarshal, we first need some JSON, so let's marshal...
		d = NewNamespacesDict()
		d.AllNamespaces[model.UserNS][userns.ID()] = userns
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())

		// ...now unmarshal again and see what nonsense we got...
		d2 := NewNamespacesDict()
		Expect(d2.UnmarshalJSON(j)).To(Succeed())
		uns := d2.AllNamespaces[model.UserNS][userns.ID()]
		Expect(uns).NotTo(BeNil())
		Expect(uns.Ref()).To(Equal(userns.Ref()))
		Expect(uns.LeaderPIDs()).To(Equal(userns.LeaderPIDs()))
		// We even should have a preliminary parent user namespace present...
		Expect(uns.(model.Hierarchy).Parent().(model.Namespace).ID()).To(
			Equal(userns.(model.Hierarchy).Parent().(model.Namespace).ID()))
	})

	It("survives a NamespacesDict roundtrip", func() {
		d := &NamespacesDict{ // TODO: use convenience helper?
			AllNamespaces: &allns.Namespaces,
			ProcessTable:  ProcessTable{allns.Processes, nil},
		}
		d.ProcessTable.Namespaces = d
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())
		Expect(j).NotTo(HaveLen(0))

		d2 := NewNamespacesDict()
		Expect(json.Unmarshal(j, &d2)).To(Succeed())

		allns2 := (*model.AllNamespaces)(d2.AllNamespaces)
		for idx := model.NamespaceTypeIndex(0); idx < model.NamespaceTypesCount; idx++ {
			nsset := allns.Namespaces[idx]
			Expect(allns2[idx]).To(HaveLen(len(nsset)))
			for _, ns := range allns2[idx] {
				Expect(ns).To(gmodel.BeSameNamespace(nsset[ns.ID()]))
			}
		}
	})

})
