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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstest/gmodel"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/testbasher"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/thediveo/success"

	. "github.com/thediveo/lxkns/nstest/gmodel"
)

var sleepyname = "morbid_moby" + strconv.FormatInt(GinkgoRandomSeed(), 10)

var (
	allns      *discover.Result
	scripts    = testbasher.Basher{}
	scriptscmd *testbasher.TestCommand
	userns     model.Namespace
	usernames  discover.UidUsernameMap

	cizer containerizer.Containerizer
)

var _ = BeforeSuite(func(ctx context.Context) {
	// Spin up a Docker engine watcher and wait for it to become ready...
	docksock := ""
	if os.Getegid() == 0 {
		docksock = "unix:///proc/1/root/run/docker.sock"
	}

	DeferCleanup(func() { scripts.Done() })

	By("creating a new Docker session for testing")
	sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=api.types.namespaces")))
	DeferCleanup(func(ctx context.Context) {
		sess.Close(ctx)
	})

	By("creating a canary workload")
	sleepy := Successful(sess.Run(ctx, "busybox:latest",
		run.WithName(sleepyname),
		run.WithCommand("/bin/sleep", "120s"),
		run.WithAutoRemove(),
		run.WithLabel("foo=bar")))
	// Make sure that the newly created container is in running state before we
	// run unit tests which depend on the correct list of alive(!)=running
	// containers.
	Expect(sleepy.PID(ctx)).NotTo(BeZero())

	mobywatcher, err := moby.New(docksock, nil)
	Expect(err).NotTo(HaveOccurred())

	cizerctx, cizercancel := context.WithCancel(context.Background())
	DeferCleanup(func() { cizercancel() })
	cizer = whalefriend.New(cizerctx, []watcher.Watcher{mobywatcher})
	Expect(cizer).NotTo(BeNil())
	Eventually(mobywatcher.Ready()).
		Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).Should(BeClosed())

	// Add some controlled namespaces for discovery...
	scripts.Common(nstest.NamespaceUtilsScript)
	scripts.Script("main", `
unshare -Ur unshare -U $stage2 # create a new user ns inside another user ns (so we get a proper owner relationship).
`)
	scripts.Script("stage2", `
process_namespaceid user # prints the user namespace ID of "the" process.
read # wait for test to proceed()
`)
	scriptscmd = scripts.Start("main")
	DeferCleanup(func() { scriptscmd.Close() })
	usernsid := nstest.CmdDecodeNSId(scriptscmd)

	// "nearly-all-ns" and ... containerz!
	allns = discover.Namespaces(
		discover.WithStandardDiscovery(),
		discover.NotFromFds(), discover.NotFromBindmounts(),
		discover.WithMounts(),
		discover.WithContainerizer(cizer))

	// basic checks that discovery worked as expected; we're here to test the
	// (un)marshalling, not discovery.
	userns = allns.Namespaces[model.UserNS][usernsid]
	Expect(userns).NotTo(BeNil())
	Expect(allns.Containers).To(ContainElement(
		PointTo(MatchFields(IgnoreExtras, Fields{
			"Name": Equal(sleepyname),
		}))))

	// For some JSON tests we need the names of the users owning user
	// namespaces. We're expecting here that the user name discovery has been
	// tested in its own defining module and thus use its results as-is for
	// simplicity.
	usernames = discover.DiscoverUserNames(allns.Namespaces)
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

// returns a "reference:ref," string to insert into a JSON object serialization,
// if the specified reference isn't empty.
func refifnotempty(ref model.NamespaceRef) string {
	if len(ref) == 0 {
		return ""
	}
	b, err := json.Marshal(ref)
	if err != nil {
		panic(err)
	}
	return `"reference": ` + string(b) + `,`
}

var _ = Describe("namespaces JSON", func() {

	It("always gets a Namespace from the dictionary", func() {
		d := NewNamespacesDict(nil)
		uns := namespaces.NewWithSimpleRef(species.CLONE_NEWUSER, species.NamespaceIDfromInode(123), "/foobar")
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
		d := NewNamespacesDict(nil)
		j, err := d.marshalNamespace(ns, nil)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "net",
				"owner": %d,
				"reference": [%q],
				"leaders": %s,
				"ealdorman": %d
			}`,
			ns.ID().Ino,
			ns.Owner().(model.Namespace).ID().Ino,
			ns.Ref(),
			pidlist(ns.LeaderPIDs()),
			ns.Ealdorman().PID,
		)))

		// In contrast, a user namespace must contain parent and UID
		// information. But it must not contain an owner reference, this is
		// the parent reference instead.
		parentuserns := userns.(model.Hierarchy).Parent().(model.Namespace)
		j, err = d.marshalNamespace(userns, usernames)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				%s
				"leaders": %s,
				"ealdorman": %d,
				"parent": %d,
				"user-id": %d,
				"user-name": %q
			}`,
			userns.ID().Ino,
			refifnotempty(userns.Ref()),
			pidlist(userns.LeaderPIDs()),
			userns.Ealdorman().PID,
			parentuserns.ID().Ino,
			userns.(model.Ownership).UID(),
			usernames[uint32(userns.(model.Ownership).UID())],
		)))

		// Check for the correct child list of the parent user namespace.
		j, err = d.marshalNamespace(parentuserns, usernames)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				%s
				"parent": %d,
				"children": %s,
				"user-id": %d,
				"user-name": %q
			}`,
			parentuserns.ID().Ino,
			refifnotempty(parentuserns.Ref()),
			parentuserns.(model.Hierarchy).Parent().(model.Namespace).ID().Ino,
			childlist(parentuserns.(model.Hierarchy)),
			parentuserns.(model.Ownership).UID(),
			usernames[uint32(parentuserns.(model.Ownership).UID())],
		)))

		// Also check the grandparent user namespace.
		grandpa := parentuserns.(model.Hierarchy).Parent().(model.Namespace)
		j, err = d.marshalNamespace(grandpa, usernames)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
				"nsid": %d,
				"type": "user",
				%s
				"leaders": %s,
				"ealdorman": %d,
				"children": %s,
				"user-id": %d,
				"user-name": %q
			}`,
			grandpa.ID().Ino,
			refifnotempty(grandpa.Ref()),
			pidlist(grandpa.LeaderPIDs()),
			grandpa.Ealdorman().PID,
			childlist(grandpa.(model.Hierarchy)),
			grandpa.(model.Ownership).UID(),
			usernames[uint32(grandpa.(model.Ownership).UID())],
		)))
	})

	It("unmarshals Namespace", func() {
		d := NewNamespacesDict(nil)
		// This unmarshalling MUST fail...
		_, err := d.UnmarshalNamespace([]byte(`""`))
		Expect(err).To(HaveOccurred())
		_, err = d.UnmarshalNamespace([]byte(`{"type":"foobar"}`))
		Expect(err).To(HaveOccurred())
		_, err = d.UnmarshalNamespace([]byte(`{"nsid":0,"type":"net"}`))
		Expect(err).To(HaveOccurred())

		// First create a JSON textual representation for a user namespace we
		// want to unmarshal next...
		j, err := d.marshalNamespace(userns, nil)
		Expect(err).To(Succeed())

		// ...now check that unmarshalling correctly works.
		nsdict := NewNamespacesDict(nil)
		uns, err := nsdict.UnmarshalNamespace(j)
		Expect(err).To(Succeed())
		Expect(uns).To(BeSameNamespace(userns))

		// Check that unmarshalling a (flat) namespace also works correctly.
		ns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS]
		j, err = nsdict.marshalNamespace(ns, nil)
		Expect(err).To(Succeed())

		ns2, err := nsdict.UnmarshalNamespace(j)
		Expect(err).To(Succeed())
		Expect(ns2).To(gmodel.BeSameNamespace(ns))
	})

	It("marshals NamespacesDict", func() {
		d := NewNamespacesDict(nil)
		d.AllNamespaces[model.UserNS][userns.ID()] = userns
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())
		username := regexp.MustCompile(`"user-name":"(.*?)"`).FindStringSubmatch(string(j))[1]
		Expect(j).To(MatchJSON(fmt.Sprintf(`{
			"%d": {
				"nsid": %[1]d,
				"type": "user",
				%s
				"leaders": %s,
				"ealdorman": %d,
				"parent": %d,
				"user-id": %d,
				"user-name": %q
			}
		}`,
			userns.ID().Ino,
			refifnotempty(userns.Ref()),
			pidlist(userns.LeaderPIDs()),
			userns.Ealdorman().PID,
			userns.(model.Hierarchy).Parent().(model.Namespace).ID().Ino,
			userns.(model.Ownership).UID(),
			username,
		)))
	})

	It("unmarshals NamespacesDict", func() {
		d := NewNamespacesDict(nil)
		// This must NOT succeed...
		Expect(d.UnmarshalJSON([]byte(`""`))).To(HaveOccurred())
		Expect(d.UnmarshalJSON([]byte(`{"123":{"type":"foobar"}}`))).To(HaveOccurred())

		// To unmarshal, we first need some JSON, so let's marshal...
		d = NewNamespacesDict(nil)
		d.AllNamespaces[model.UserNS][userns.ID()] = userns
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())

		// ...now unmarshal again and see what nonsense we got...
		d2 := NewNamespacesDict(nil)
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
		d := NewNamespacesDict(allns)
		j, err := json.Marshal(d)
		Expect(err).To(Succeed())
		Expect(j).NotTo(HaveLen(0))

		d2 := NewNamespacesDict(nil)
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
