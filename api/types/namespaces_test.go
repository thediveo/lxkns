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
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
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

	It("marshals Namespace", func() {
		// A non-user and non-PID namespace must not contain parent and UID
		// information. But it must contain a owner user namespace reference.
		ns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS]
		Expect(ns).NotTo(BeNil())
		j, err := MarshalNamespace(ns)
		Expect(err).NotTo(HaveOccurred())
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
		j, err = MarshalNamespace(userns)
		Expect(err).NotTo(HaveOccurred())
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
		j, err = MarshalNamespace(parentuserns)
		Expect(err).NotTo(HaveOccurred())
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
		j, err = MarshalNamespace(grandpa)
		Expect(err).NotTo(HaveOccurred())
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
		// This unmarshalling MUST fail...
		_, err := UnmarshalNamespace([]byte(`""`), nil, nil)
		Expect(err).To(HaveOccurred())
		_, err = UnmarshalNamespace([]byte(`{"type":"foobar"}`), nil, nil)
		Expect(err).To(HaveOccurred())
		_, err = UnmarshalNamespace([]byte(`{"nsid":0,"type":"net"}`), nil, nil)
		Expect(err).To(HaveOccurred())

		// First create a JSON textual representation for a user namespace we
		// want to unmarshal next...
		j, err := MarshalNamespace(userns)
		Expect(err).NotTo(HaveOccurred())

		// ...now check that unmarshalling correctly works.
		nsdict := NewNamespacesDict()
		procs := model.ProcessTable{}
		uns, err := UnmarshalNamespace(j, nsdict, procs)
		Expect(err).NotTo(HaveOccurred())
		Expect(uns).NotTo(BeNil())
		Expect(uns.ID()).To(Equal(userns.ID()))
		Expect(uns.Type()).To(Equal(userns.Type()))
		Expect(uns.Ref()).To(Equal(userns.Ref()))

		Expect(uns.(model.Hierarchy).Parent().(model.Namespace).ID()).To(
			Equal(userns.(model.Hierarchy).Parent().(model.Namespace).ID()))
		Expect(uns.LeaderPIDs()).To(Equal(userns.LeaderPIDs()))

		// Check that unmarshalling a (flat) namespace also works correctly.
		ns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS]
		j, err = MarshalNamespace(ns)
		Expect(err).NotTo(HaveOccurred())

		ns2, err := UnmarshalNamespace(j, nsdict, procs)
		Expect(err).NotTo(HaveOccurred())
		Expect(ns2).NotTo(BeNil())
		Expect(ns2.ID()).To(Equal(ns.ID()))
		Expect(ns2.Type()).To(Equal(ns.Type()))
		Expect(ns2.Ref()).To(Equal(ns.Ref()))
		Expect(ns2.Owner().(model.Namespace).ID()).To(Equal(ns.Owner().(model.Namespace).ID()))
	})

})
