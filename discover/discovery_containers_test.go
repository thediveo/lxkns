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

//go:build linux

package discover

import (
	"context"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var sleepyname = "dumb_doormat" + strconv.FormatInt(GinkgoRandomSeed(), 10)

var noDockerRE = regexp.MustCompile(`connect: no such file or directory`)

var _ = Describe("Discover containers", func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	var sleepy *morbyd.Container
	var docksock string

	BeforeEach(func(ctx context.Context) {
		// We cannot discover the initial container process running as root when
		// we're not root too.
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		docksock = "unix:///proc/1/root/run/docker.sock"

		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=discover.containers")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating canary workloads")
		sleepy = Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(sleepyname),
			run.WithCommand("/bin/sleep", "120s"),
			run.WithAutoRemove(),
			run.WithLabel(composer.ComposerProjectLabel+"=lxkns-project")))
		// Make sure that the newly created container is in running state before we
		// run unit tests which depend on the correct list of alive(!)=running
		// containers.
		Expect(sleepy.PID(ctx)).NotTo(BeZero())
	})

	It("finds containers and relates them with their initial processes", func() {
		By("spinning up a Docker watcher")
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()
		Eventually(mw.Ready()).Should(BeClosed(), "dockerd watcher failed to synchronize")

		By("looking for the sleepy container")
		allns := Namespaces(WithStandardDiscovery(), WithContainerizer(cizer))
		sleepy := allns.Containers.FirstWithName(sleepyname)
		Expect(sleepy).NotTo(BeNil())
		Expect(sleepy.PID).NotTo(BeZero())
		Expect(allns.Processes).To(HaveKey(sleepy.PID))
		Expect(sleepy.Process).NotTo(BeNil())
		Expect(sleepy.Process.Container).To(Equal(sleepy))

		g := sleepy.Group(composer.ComposerGroupType)
		Expect(g).NotTo(BeNil())
		Expect(g.Type).To(Equal(composer.ComposerGroupType))
		Expect(g.Name).To(Equal("lxkns-project"))
		Expect(g.Containers).To(HaveLen(1))
		Expect(g.Containers).To(ConsistOf(sleepy))
	})

})
