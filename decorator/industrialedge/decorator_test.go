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

package industrialedge

import (
	"context"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

var names = map[string]struct {
	projectname string
}{
	"edgy_emil":              {projectname: "foobar_project"},
	"furious_freddy":         {projectname: "foobar_project"},
	edgeRuntimeContainerName: {},
}

var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

var _ = Describe("Decorates composer projects", func() {

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
		for name, config := range names {
			_ = pool.RemoveContainerByName(name)
			var labels map[string]string
			if config.projectname != "" {
				labels = map[string]string{
					composer.ComposerProjectLabel:    config.projectname,
					edgeAppConfigLabelPrefix + "foo": "bar",
				}
			}
			sleepy, err := pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       name,
				Cmd:        []string{"/bin/sleep", "120s"},
				Labels:     labels,
			})
			// Skip test in case Docker is not accessible.
			if err != nil && nodockerre.MatchString(err.Error()) {
				Skip("Docker not available")
			}
			Expect(err).NotTo(HaveOccurred(), "container %q", name)
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

	It("decorates IE apps and IED runtime", func() {
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())
		composer.Decorate([]*model.ContainerEngine{allcontainers[0].Engine}, nil)
		Decorate([]*model.ContainerEngine{allcontainers[0].Engine}, nil)

		containers := make([]*model.Container, 0, len(names))
		for _, container := range allcontainers {
			if _, ok := names[container.Name]; ok {
				containers = append(containers, container)
			}
		}
		Expect(containers).To(HaveLen(len(names)))

		Expect(containers).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
			"Name":   Equal(edgeRuntimeContainerName),
			"Flavor": Equal(IndustrialEdgeRuntimeFlavor),
		}))))

		for _, container := range containers {
			if names[container.Name].projectname == "" {
				continue
			}
			g := container.Group(composer.ComposerGroupType)
			Expect(g).NotTo(BeNil())
			Expect(g.Type).To(Equal(composer.ComposerGroupType))
			Expect(g.Flavor).To(Equal(IndustrialEdgeAppFlavor))
		}
	})

})
