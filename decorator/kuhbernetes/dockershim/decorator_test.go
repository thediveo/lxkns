// Copyright 2021 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockershim

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/matcher"
	. "github.com/thediveo/success"
)

var _ = Describe("Decorates k8s docker shim containers", Ordered, func() {

	var names = map[string]bool{
		"k8s_foo_foopod_foons_123uid_1": true,
		"k8s_bar_foopod_foons_123uid_1": true,
		"k8s_POD_foopod_foons_123uid_1": true,
		"k8s_123":                       false,
		"k8s_abc-_def_ghi_123_1":        false,
		"k8s_abc_def-_ghi_123_1":        false,
		"k8s_abc_def_ghi-_123_1":        false,
	}

	var sleepies []*morbyd.Container

	BeforeAll(slowSpec, func(ctx context.Context) {
		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=decorator.kuhbernetes")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating fake pod containers")
		for name := range names {
			sleepy := Successful(sess.Run(ctx, "busybox:latest",
				run.WithName(name),
				run.WithCommand("/bin/sleep", "120s"),
				run.WithAutoRemove()))
			// Make sure that the newly created container is in running state before we
			// run unit tests which depend on the correct list of alive(!)=running
			// containers.
			Expect(sleepy.PID(ctx)).NotTo(BeZero())
			sleepies = append(sleepies, sleepy)
		}
	})

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		ignoreGood := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked(ignoreGood))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("decorates k8s pods", slowSpec, func(ctx context.Context) {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		mw, err := moby.New("", nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())
		Decorate([]*model.ContainerEngine{allcontainers[0].Engine}, nil)

		By("checking for all fakes to be found")
		containers := make([]*model.Container, 0, len(names))
		for _, container := range allcontainers {
			if _, ok := names[container.Name]; ok {
				containers = append(containers, container)
			}
		}
		Expect(containers).To(HaveLen(len(names)))

		By("checking for correct decoration")
		for _, container := range containers {
			g := container.Group(kuhbernetes.PodGroupType)
			if !names[container.Name] {
				Expect(g).To(BeNil(), "non-pod container %s", container.Name)
				continue
			}
			Expect(g).To(BeAPod(), "pod container %s", container.Name)
			Expect(g.Containers).To(ContainElement(container), container.Name)
			Expect(container.Labels[kuhbernetes.PodUidLabel]).To(Equal("123uid"), container.Name)
			if strings.Contains(container.Name, "_POD_") {
				Expect(container).To(HaveLabel(kuhbernetes.PodSandboxLabel), container.Name)
			} else {
				Expect(container).NotTo(HaveLabel(kuhbernetes.PodSandboxLabel), container.Name)
			}
		}
	})

})
