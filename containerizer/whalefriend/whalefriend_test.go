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
	"os"
	"regexp"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/lxkns/matcher"
)

const sleepyname = "pompous_pm"

var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

var _ = Describe("ContainerEngine", func() {

	var pool *dockertest.Pool
	var sleepy *dockertest.Resource
	var docksock string

	BeforeEach(func() {
		// In case we're run as root we use a procfs wormhole so we can access
		// the Docker socket even from a test container without mounting it
		// explicitly into the test container.
		if os.Geteuid() == 0 {
			docksock = "unix:///proc/1/root/run/docker.sock"
		}

		var err error
		pool, err = dockertest.NewPool(docksock)
		Expect(err).NotTo(HaveOccurred())
		_ = pool.RemoveContainerByName(sleepyname)
		sleepy, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       sleepyname,
			Cmd:        []string{"/bin/sleep", "120s"},
			Labels:     map[string]string{"foo": "bar"},
		})
		// Skip test in case Docker is not accessible.
		if err != nil && nodockerre.MatchString(err.Error()) {
			Skip("Docker not available")
		}
		Expect(err).NotTo(HaveOccurred(), "container %q", sleepyname)
		Eventually(func() bool {
			c, err := pool.Client.InspectContainer(sleepy.Container.ID)
			Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
			return c.State.Running
		}, "5s", "100ms").Should(BeTrue(), "container %s", sleepy.Container.Name[1:])
		Expect(pool.Client.PauseContainer(sleepyname)).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(pool.Client.UnpauseContainer(sleepyname)).NotTo(HaveOccurred())
		Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
	})

	It("discovers container", func() {
		dockerw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cew := New(ctx, []watcher.Watcher{dockerw})
		Expect(cew).NotTo(BeNil())

		// wait for the watcher to have completed its initial synchronization
		// with its container engine...
		Eventually(dockerw.Ready(), "5s", "100ms").Should(BeClosed())
		// ...then wait for it to have also picked up the paused state of our
		// test container (better safe than sorry in this case).
		Eventually(func() bool {
			return dockerw.Portfolio().Container(sleepyname).Paused
		}).Should(BeTrue())

		cntrs := cew.Containers(ctx, nil, nil)
		Expect(cntrs).To(ContainElement(HaveContainerName(sleepyname)))

		var c *model.Container
		for _, cntr := range cntrs {
			if cntr.Name == sleepyname {
				c = cntr
				break
			}
		}
		// dockertest's resource does not properly reflect container state
		// changes after creation, sigh. So we need to inspect to get the
		// correct information.
		csleepy, err := pool.Client.InspectContainer(sleepy.Container.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.ID).To(Equal(csleepy.ID))
		Expect(c.PID).To(Equal(model.PIDType(csleepy.State.Pid)))
		Expect(c.Paused).To(Equal(csleepy.State.Paused))
		Expect(c.Labels).To(HaveKeyWithValue("foo", "bar"))
		Expect(c.Type).To(Equal(moby.Type))

		e := c.Engine
		Expect(e).NotTo(BeNil())
		Expect(e.Type).To(Equal(moby.Type))
		Expect(e.API).NotTo(BeEmpty())
		Expect(e.ID).NotTo(BeEmpty())
	})

})
