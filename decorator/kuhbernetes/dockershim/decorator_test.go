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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ory/dockertest"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

var names = map[string]bool{
	"k8s_foo_foopod_foons_123uid_1": true,
	"k8s_bar_foopod_foons_123uid_1": true,
	"k8s_POD_foopod_foons_123uid_1": true,
	"k8s_123":                       false,
	"k8s_abc-_def_ghi_123_1":        false,
	"k8s_abc_def-_ghi_123_1":        false,
	"k8s_abc_def_ghi-_123_1":        false,
}

var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

var _ = Describe("Decorates k8s docker shim containers", func() {

	var pool *dockertest.Pool
	var sleepies []*dockertest.Resource
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
		for name := range names {
			_ = pool.RemoveContainerByName(name)
			Eventually(func() error {
				_, err := pool.Client.InspectContainer(name)
				return err
			}, "5s", "100ms").Should(HaveOccurred())
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
		// Make sure that all newly created containers are in running state
		// before we run unit tests which depend on the correct list of
		// alive(!)=running containers.
		for _, sleepy := range sleepies {
			Eventually(func() bool {
				c, err := pool.Client.InspectContainer(sleepy.Container.ID)
				Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
				return c.State.Running
			}, "5s", "100ms").Should(BeTrue(), "container %s", sleepy.Container.Name[1:])
		}
	})

	AfterEach(func() {
		for _, sleepy := range sleepies {
			Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
		}
	})

	It("decorates k8s pods", func() {
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())
		Decorate([]*model.ContainerEngine{allcontainers[0].Engine})

		containers := make([]*model.Container, 0, len(names))
		for _, container := range allcontainers {
			if _, ok := names[container.Name]; ok {
				containers = append(containers, container)
			}
		}
		Expect(containers).To(HaveLen(len(names)))

		for _, container := range containers {
			g := container.Group(kuhbernetes.PodGroupType)
			if !names[container.Name] {
				Expect(g).To(BeNil(), "non-pod container %s", container.Name)
				continue
			}
			Expect(g).NotTo(BeNil(), "pod container %s", container.Name)
			Expect(g.Type).To(Equal(kuhbernetes.PodGroupType), container.Name)
			Expect(g.Containers).To(ContainElement(container), container.Name)
			Expect(container.Labels[kuhbernetes.PodUidLabel]).To(Equal("123uid"), container.Name)
			if strings.Contains(container.Name, "_POD_") {
				Expect(container.Labels).To(HaveKey(kuhbernetes.PodSandboxLabel), container.Name)
			} else {
				Expect(container.Labels).NotTo(HaveKey(kuhbernetes.PodSandboxLabel), container.Name)
			}
		}
	})

})
