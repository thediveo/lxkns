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

	"github.com/ory/dockertest"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const sleepyname = "pompous_pm"

var _ = Describe("ContainerEngine", func() {

	var pool *dockertest.Pool
	var sleepy *dockertest.Resource

	BeforeEach(func() {
		var err error
		pool, err = dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		sleepy, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       sleepyname,
			Cmd:        []string{"/bin/sleep", "30s"},
			Labels:     map[string]string{"foo": "bar"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(pool.Client.PauseContainer(sleepyname)).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(pool.Client.UnpauseContainer(sleepyname)).NotTo(HaveOccurred())
		Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
	})

	It("discovers container", func() {
		dockerw, err := moby.NewWatcher("")
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cew := New(ctx, []watcher.Watcher{dockerw})
		Expect(cew).NotTo(BeNil())

		// wait for the watcher to have completed its initial synchronization
		// with its container engine...
		<-dockerw.Ready()
		// ...then wait for it to have also picked up the paused state of our
		// test container (better safe than sorry in this case).
		Eventually(func() bool {
			return dockerw.Portfolio().Container(sleepyname).Paused
		}).Should(BeTrue())

		cntrs := cew.Containers(ctx, nil, nil)
		Expect(cntrs).To(ContainElement(
			WithTransform(func(c model.Container) string { return c.Name() }, Equal(sleepyname))))

		var c model.Container
		for _, cntr := range cntrs {
			if cntr.Name() == sleepyname {
				c = cntr
				break
			}
		}
		// dockertest's resource does not properly reflect container state
		// changes after creation, sigh. So we need to inspect to get the
		// correct information.
		csleepy, err := pool.Client.InspectContainer(sleepy.Container.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.ID()).To(Equal(csleepy.ID))
		Expect(c.PID()).To(Equal(model.PIDType(csleepy.State.Pid)))
		Expect(c.Paused()).To(Equal(csleepy.State.Paused))
		Expect(c.Labels()).To(HaveKeyWithValue("foo", "bar"))
		Expect(c.Type()).To(Equal(moby.Type))
		Expect(c.Engine()).NotTo(BeNil())
		Expect(c.Engine().API()).NotTo(BeEmpty())
		Expect(c.Engine().ID()).NotTo(BeEmpty())
	})

})
