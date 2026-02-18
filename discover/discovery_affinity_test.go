// Copyright 2026 Harald Albrecht.
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

package discover

import (
	"log/slog"
	"time"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("Discover affinities", func() {

	BeforeEach(func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(func(olddepth uint) { format.MaxDepth = olddepth }, format.MaxDepth)
		format.MaxDepth = 2
	})

	It("discovers process affinities and online CPUs", func() {
		allns := Namespaces(FromProcs(), WithAffinityAndScheduling())
		Expect(allns.OnlineCPUs).NotTo(BeEmpty())
		Expect(allns.Processes).To(ContainElement(HaveField("Affinity", Not(BeEmpty()))))
	})

	It("discovers task and affinities, as well as online CPUs", func() {
		allns := Namespaces(FromTasks(), WithTaskAffinityAndScheduling())
		Expect(allns.OnlineCPUs).NotTo(BeEmpty())
		Expect(allns.Processes).To(ContainElement(HaveField("Affinity", Not(BeEmpty()))))
		Expect(allns.Processes).To(ContainElement(
			HaveField("Tasks", ContainElement(HaveField("Affinity", Not(BeEmpty()))))))
	})

})
