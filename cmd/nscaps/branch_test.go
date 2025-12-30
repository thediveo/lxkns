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

package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

// GomegaString returns a simplified representation of this namespace node, in
// order to avoid Gomega's very detailed object dumps when dealing with
// branches.
func (n namespaceNode) GomegaString() string {
	var s strings.Builder

	s.WriteString("namespaceNode: ")

	if n.isTarget {
		s.WriteString("🞋")
	}

	s.WriteString(n.ns.Type().Name())
	s.WriteString(":[")
	s.WriteString(strconv.FormatUint(n.ns.ID().Ino, 10))
	s.WriteString("] ")

	s.WriteString(nscapmarks[n.targetCapsSummary])

	if len(n.children) == 0 {
		return s.String()
	}

	for _, child := range n.children {
		s.WriteRune('\n')
		s.WriteString(format.IndentString(child.(format.GomegaStringer).GomegaString(), 1))
	}

	return s.String()
}

// GomegaString returns a simplified representation of this process, in order to
// avoid Gomega's overly details object dumps when dealing with branches.
func (p processNode) GomegaString() string {
	var s strings.Builder

	s.WriteString("processNode: ")
	s.WriteString(p.proc.Name)
	s.WriteRune('(')
	s.WriteString(strconv.FormatUint(uint64(p.proc.PID), 10))
	s.WriteRune(')')

	return s.String()
}

func last(n node) node {
	Expect(n).WithOffset(1).NotTo(BeNil())
	Expect(n).WithOffset(1).To(BeAssignableToTypeOf(&namespaceNode{}))
	for len(n.Children()) == 1 {
		Expect(n).WithOffset(1).To(BeAssignableToTypeOf(&namespaceNode{}))
		n = n.Children()[0]
	}
	return n
}

var _ = Describe("user/PID branches and capabilities", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		// As we're keeping a harness script running in the background we'll
		// have additionally "background" goroutines running that would
		// otherwise cause false positives, so we take a snapshot here.
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	Context("(in)capability", func() {

		It("get euid for arbitrary process", func() {
			euid := os.Geteuid()
			Expect(processEuid(&model.Process{PID: model.PIDType(os.Getpid())})).To(Equal(euid))
		})

		It("fails correctly", func() {
			_, _, err := caps(&model.Process{PID: 1}, nil)
			Expect(err).To(HaveOccurred())
			_, _, err = caps(allns.Processes[model.PIDType(os.Getpid())],
				namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NoneID, ""))
			Expect(err).To(HaveOccurred())
		})

		It("rule #1: process inside target's user namespace", func() {
			tcaps, euid, err := caps(
				allns.Processes[targetPID],
				allns.Namespaces[model.UserNS][targetUsernsID])
			Expect(err).ToNot(HaveOccurred())
			Expect(tcaps).To(Equal(effcaps))
			Expect(euid).To(Equal(processEuid(allns.Processes[targetPID])))

		})

		It("missed rule #2: process inside target's sibling user namespace", func() {
			tcaps, _, err := caps(
				allns.Processes[someProcPID],
				allns.Namespaces[model.NetNS][targetNetnsID])
			Expect(err).ToNot(HaveOccurred())
			Expect(tcaps).To(Equal(incapable))
		})

		It("rules #2+#3: this process outside target's user namespace hierarchy", func() {
			tcaps, euid, err := caps(
				allns.Processes[model.PIDType(os.Getpid())],
				allns.Namespaces[model.NetNS][targetNetnsID])
			Expect(err).ToNot(HaveOccurred())
			Expect(tcaps).To(Equal(allcaps))
			Expect(euid).To(Equal(os.Geteuid()))
		})

		It("rule #2, but not #3: process inside target's user namespace hierarchy, but different owner", func() {
			if os.Geteuid() == 0 {
				Skip("only non-root")
			}
			// need to fake this in case we're not ro(o)t.
			fakeinit := &model.Process{PID: 1}
			fakeinit.Namespaces[model.UserNS] =
				allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.UserNS]
			tcaps, _, err := caps(fakeinit, allns.Namespaces[model.NetNS][targetNetnsID])
			Expect(err).ToNot(HaveOccurred())
			Expect(tcaps).To(Equal(effcaps))
		})

	})

	Context("branches and forks", func() {

		It("creates process branch", func() {
			_, err := branchOfProcess(&model.Process{PID: -1}, 0)
			Expect(err).To(HaveOccurred())

			euid := os.Getegid()
			procbr, err := branchOfProcess(allns.Processes[targetPID], euid)
			Expect(err).ToNot(HaveOccurred())
			l := last(procbr)
			Expect(l).To(BeAssignableToTypeOf(&processNode{}))
			Expect(l.(*processNode).proc.PID).To(Equal(targetPID))
			Expect(l.(*processNode).euid).To(Equal(euid))
		})

		It("creates target branch for user namespace", func() {
			tbr := branchOfTarget(allns.Namespaces[model.UserNS][targetUsernsID], effcaps)
			Expect(tbr).ToNot(BeNil())
			l := last(tbr)
			Expect(l).To(BeAssignableToTypeOf(&namespaceNode{}))
			Expect(l.(*namespaceNode).isTarget).To(BeTrue())
			Expect(l.(*namespaceNode).targetCapsSummary).To(Equal(effcaps))
			Expect(l.(*namespaceNode).ns.ID()).To(Equal(targetUsernsID))
		})

		It("creates target branch for non-user namespace", func() {
			tbr := branchOfTarget(allns.Namespaces[model.NetNS][targetNetnsID], incapable)
			Expect(tbr).ToNot(BeNil())
			l := last(tbr)
			Expect(l).To(BeAssignableToTypeOf(&namespaceNode{}))
			Expect(l.(*namespaceNode).isTarget).To(BeTrue())
			Expect(l.(*namespaceNode).ns.ID()).To(Equal(targetNetnsID))
		})

		When("combining branches", func() {

			It("process outside target hierarchy", func() {
				procbr, err := branchOfProcess(allns.Processes[model.PIDType(os.Getpid())], 0)
				Expect(err).ToNot(HaveOccurred())
				tbr := branchOfTarget(allns.Namespaces[model.NetNS][targetNetnsID], incapable)
				root := combine(procbr, tbr)
				Expect(root).NotTo(BeNil())
				Expect(root.Children()).To(HaveLen(2))
				Expect(root.Children()[0]).To(BeAssignableToTypeOf(&processNode{}))
				Expect(root.Children()[1]).To(BeAssignableToTypeOf(&namespaceNode{}))
				Expect(root.Children()[1].(*namespaceNode).ns.ID()).To(Equal(targetUsernsID))
			})

			It("process inside target hierarchy", func() {
				procbr, err := branchOfProcess(allns.Processes[targetPID], 0)
				Expect(err).ToNot(HaveOccurred())
				tbr := branchOfTarget(allns.Namespaces[model.NetNS][targetNetnsID], incapable)
				root := combine(procbr, tbr)
				Expect(root).NotTo(BeNil())
				Expect(root.Children()).To(HaveLen(1))
				userchild := root.Children()[0]

				Expect(userchild).To(BeAssignableToTypeOf(&namespaceNode{}))
				Expect(userchild.(*namespaceNode).ns.ID()).To(Equal(targetUsernsID))
				Expect(userchild.Children()).To(HaveLen(2))
				Expect(userchild.Children()[0]).To(BeAssignableToTypeOf(&processNode{}))
				Expect(userchild.Children()[1]).To(BeAssignableToTypeOf(&namespaceNode{}))
				Expect(userchild.Children()[1].(*namespaceNode).ns.ID()).To(Equal(targetNetnsID))
			})

			It("process outside target hierarchy in another child user namespace", func() {
				procbr, err := branchOfProcess(allns.Processes[someProcPID], 0)
				Expect(err).ToNot(HaveOccurred())
				tbr := branchOfTarget(allns.Namespaces[model.NetNS][targetNetnsID], incapable)
				root := combine(procbr, tbr)
				Expect(root).NotTo(BeNil())
				Expect(root.Children()).To(HaveLen(2))

				procuserchild := root.Children()[0]
				Expect(procuserchild).To(BeAssignableToTypeOf(&namespaceNode{}))
				Expect(procuserchild.Children()).To(HaveLen(1))
				Expect(procuserchild.Children()[0]).To(BeAssignableToTypeOf(&processNode{}))

				userchild := root.Children()[1]
				Expect(userchild).To(BeAssignableToTypeOf(&namespaceNode{}))
				Expect(userchild.Children()[0].(*namespaceNode).ns.ID()).To(Equal(targetNetnsID))
			})

			It("process below target namespace", func() {
				procbr, err := branchOfProcess(allns.Processes[targetPID], 0)
				Expect(err).ToNot(HaveOccurred())
				tbr := branchOfTarget(allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.UserNS], incapable)
				root := combine(procbr, tbr)
				Expect(root).NotTo(BeNil())
				Expect(root.(*namespaceNode).isTarget).To(BeTrue())
				Expect(root.Children()).To(HaveLen(1))

				userchild := root.Children()[0]
				Expect(userchild).To(BeAssignableToTypeOf(&namespaceNode{}))

				Expect(userchild.Children()).To(HaveLen(1))
				procchild := userchild.Children()[0]
				Expect(procchild).To(BeAssignableToTypeOf(&processNode{}))
			})

		})

	})

})
