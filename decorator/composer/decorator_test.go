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
	"regexp"
	"strconv"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/matcher"
	. "github.com/thediveo/noleak"
)

var _ = Describe("Decorates composer projects", Ordered, func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	var names = map[string]struct{}{
		"dumb_doormat" + strconv.FormatInt(GinkgoRandomSeed(), 10): {},
		"pompous_pm" + strconv.FormatInt(GinkgoRandomSeed(), 10):   {},
	}

	var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

	var pool *dockertest.Pool
	var sleepies []*dockertest.Resource
	var docksock string

	BeforeAll(func() {
		// In case we're run as root we use a procfs wormhole so we can access
		// the Docker socket even from a test container without mounting it
		// explicitly into the test container.
		if os.Geteuid() == 0 {
			docksock = "unix:///proc/1/root/run/docker.sock"
		}

		var err error
		pool, err = dockertest.NewPool(docksock)
		Expect(err).NotTo(HaveOccurred())
		for name := range names {
			_ = pool.RemoveContainerByName(name)
			sleepy, err := pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       name,
				Cmd:        []string{"/bin/sleep", "120s"},
				Labels: map[string]string{
					ComposerProjectLabel: name + "-project",
				},
			})
			// Skip test in case Docker is not accessible.
			if err != nil && nodockerre.MatchString(err.Error()) {
				Skip("Docker not available")
			}
			Expect(err).NotTo(HaveOccurred(), "container %s", name)
			sleepies = append(sleepies, sleepy)
		}
		// Make sure that all newly created containers are in running state
		// before we run unit tests which depend on the correct list of
		// alive(!)=running containers.
		for _, sleepy := range sleepies {
			Eventually(func() bool {
				c, err := pool.Client.InspectContainer(sleepy.Container.ID)
				Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
				return c.State.Running
			}).WithTimeout(5*time.Second).WithPolling(100*time.Millisecond).
				Should(BeTrue(), "container %s", sleepy.Container.Name[1:])
		}

		DeferCleanup(func() {
			for _, sleepy := range sleepies {
				Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
			}
			pool.Client.HTTPClient.CloseIdleConnections()
		})
	})

	It("decorates composer projects", func() {
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())
		Decorate([]*model.ContainerEngine{allcontainers[0].Engine}, nil)

		containers := make([]*model.Container, 0, len(names))
		for _, container := range allcontainers {
			if _, ok := names[container.Name]; ok {
				containers = append(containers, container)
			}
		}
		Expect(containers).To(HaveLen(len(names)))

		for _, container := range containers {
			Expect(container).To(BeInAGroup(WithType(ComposerGroupType)))
			Expect(container.Group(ComposerGroupType).Containers).To(ConsistOf(container))
		}
	})

})
