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
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/matcher"
)

var _ = Describe("Decorates k8s docker shim containers", Ordered, func() {

	var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

	var names = map[string]bool{
		"k8s_foo_foopod_foons_123uid_1": true,
		"k8s_bar_foopod_foons_123uid_1": true,
		"k8s_POD_foopod_foons_123uid_1": true,
		"k8s_123":                       false,
		"k8s_abc-_def_ghi_123_1":        false,
		"k8s_abc_def-_ghi_123_1":        false,
		"k8s_abc_def_ghi-_123_1":        false,
	}

	var pool *dockertest.Pool
	var sleepies []*dockertest.Resource
	var docksock string

	BeforeAll(slowSpec, func(ctx context.Context) {
		// In case we're run as root we use a procfs wormhole so we can access
		// the Docker socket even from a test container without mounting it
		// explicitly into the test container.
		if os.Geteuid() == 0 {
			docksock = "unix:///proc/1/root/run/docker.sock"
		}

		var err error
		pool, err = dockertest.NewPool(docksock)
		Expect(err).NotTo(HaveOccurred())

		By("creating fake pod containers")
		for name := range names {
			_ = pool.RemoveContainerByName(name)
			Eventually(func() error {
				_, err := pool.Client.InspectContainer(name)
				return err
			}).WithContext(ctx).Within(5 * time.Second).ProbeEvery(100 * time.Millisecond).
				Should(HaveOccurred())
			sleepy, err := pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       name,
				Cmd:        []string{"/bin/sleep", "120s"},
				Labels:     map[string]string{},
			})
			// Skip test in case Docker is not accessible.
			if err != nil && nodockerre.MatchString(err.Error()) {
				Skip("Docker not available")
			}
			Expect(err).NotTo(HaveOccurred())
			sleepies = append(sleepies, sleepy)
		}

		By("waiting for all fake containers to be running")
		// Make sure that all newly created containers are in running state
		// before we run unit tests which depend on the correct list of
		// alive(!)=running containers.
		for _, sleepy := range sleepies {
			Eventually(func() bool {
				c, err := pool.Client.InspectContainer(sleepy.Container.ID)
				Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
				return c.State.Running
			}).WithContext(ctx).Within(5*time.Second).ProbeEvery(100*time.Millisecond).
				Should(BeTrue(), "container %s", sleepy.Container.Name[1:])
		}

		DeferCleanup(func() {
			for _, sleepy := range sleepies {
				Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
			}
		})

		// Settle down things before proceeding with the real unit test.
		pool.Client.HTTPClient.CloseIdleConnections()
		Eventually(Goroutines).WithContext(ctx).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
			ShouldNot(HaveLeaked())
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
		mw, err := moby.New(docksock, nil)
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
