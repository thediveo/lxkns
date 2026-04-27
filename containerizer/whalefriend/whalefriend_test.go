// Copyright 2021 Harald Albrecht.
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

package whalefriend

import (
	"context"
	"time"

	"github.com/moby/moby/client"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/whalewatcher/v2/watcher"
	"github.com/thediveo/whalewatcher/v2/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/matcher"
	. "github.com/thediveo/success"
)

const sleepyname = "pompous_pm"

var _ = Describe("ContainerEngine", func() {

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

	var sleepy *morbyd.Container

	BeforeEach(func(ctx context.Context) {
		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=containerizer.whalefriend")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating a canary workload")
		sleepy = Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(sleepyname),
			run.WithCommand("/bin/sleep", "120s"),
			run.WithAutoRemove(),
			run.WithLabel("foo=bar")))
		// Make sure that the newly created container is in running state before we
		// run unit tests which depend on the correct list of alive(!)=running
		// containers.
		Expect(sleepy.PID(ctx)).NotTo(BeZero())

		Expect(sess.Client().ContainerPause(ctx, sleepyname, client.ContainerPauseOptions{})).Error().To(Succeed())

		DeferCleanup(func(ctx context.Context) {
			Expect(sess.Client().ContainerUnpause(ctx, sleepyname, client.ContainerUnpauseOptions{})).Error().To(Succeed())
		})
	})

	It("discovers container", func() {
		dockerw, err := moby.New("", nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cew := New(ctx, []watcher.Watcher{dockerw})
		Expect(cew).NotTo(BeNil())
		defer cew.Close()

		// wait for the watcher to have completed its initial synchronization
		// with its container engine...
		Eventually(dockerw.Ready(), "5s", "100ms").Should(BeClosed())
		// ...then wait for it to have also picked up the paused state of our
		// test container (better safe than sorry in this case).
		Eventually(func() bool {
			return dockerw.Portfolio().Container(sleepyname).Paused
		}).Should(BeTrue())

		cntrs := cew.Containers(ctx, nil, nil)
		var c *model.Container
		Expect(cntrs).To(ContainElement(WithName(sleepyname), &c))

		Expect(sleepy.Refresh(ctx)).To(Succeed())
		Expect(c.ID).To(Equal(sleepy.ID))
		Expect(c.PID).To(Equal(model.PIDType(sleepy.Details.Container.State.Pid)))
		Expect(c.Paused).To(Equal(sleepy.Details.Container.State.Paused))
		Expect(c.Labels).To(HaveKeyWithValue("foo", "bar"))
		Expect(c.Type).To(Equal(moby.Type))

		e := c.Engine
		Expect(e).NotTo(BeNil())
		Expect(e.Type).To(Equal(moby.Type))
		Expect(e.API).NotTo(BeEmpty())
		Expect(e.ID).NotTo(BeEmpty())
	})

})
