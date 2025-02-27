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

package composer

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/test/matcher"
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

var _ = Describe("Decorates composer projects", Ordered, func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	var names = []string{
		"dumb_doormat" + strconv.FormatInt(GinkgoRandomSeed(), 10),
		"pompous_pm" + strconv.FormatInt(GinkgoRandomSeed(), 10),
	}

	var sleepies []*morbyd.Container
	var docksock string

	BeforeAll(func(ctx context.Context) {
		// In case we're run as root we use a procfs wormhole so we can access
		// the Docker socket even from a test container without mounting it
		// explicitly into the test container.
		if os.Geteuid() == 0 {
			docksock = "unix:///proc/1/root/run/docker.sock"
		}

		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=decorator.composer")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating canary workloads")
		for _, name := range names {
			sleepy := Successful(sess.Run(ctx, "busybox:latest",
				run.WithName(name),
				run.WithCommand("/bin/sleep", "120s"),
				run.WithAutoRemove(),
				run.WithLabel(ComposerProjectLabel+"="+name+"-project")))
			// Make sure that the newly created container is in running state before we
			// run unit tests which depend on the correct list of alive(!)=running
			// containers.
			Expect(sleepy.PID(ctx)).NotTo(BeZero())
			sleepies = append(sleepies, sleepy)
		}
	})

	It("decorates composer projects", func() {
		By("watcher whales")
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		By("finding a Docker engine")
		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())
		var canaries []*model.Container
		Expect(allcontainers).To(ContainElement(matcher.WithType(moby.Type), &canaries))
		Expect(len(canaries)).To(BeNumerically(">=", len(names)))

		By("decorating the Docker containers")
		Decorate([]*model.ContainerEngine{canaries[0].Engine}, nil)

		var containers []*model.Container
		Expect(allcontainers).To(ContainElement(
			HaveField("Name", BeElementOf(names)),
			&containers))
		Expect(containers).To(HaveLen(len(names)), "missing canaries: %#v", allcontainers)

		for _, container := range containers {
			Expect(container).To(BeInAGroup(WithType(ComposerGroupType)))
			Expect(container.Group(ComposerGroupType).Containers).To(ConsistOf(container))
		}
	})

})
