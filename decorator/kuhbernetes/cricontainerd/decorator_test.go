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

package cricontainerd

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/containerd/containerd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/test/containerdtest"
	"github.com/thediveo/whalewatcher/watcher"
	cdwatcher "github.com/thediveo/whalewatcher/watcher/containerd"
)

var names = map[string]struct {
	labels map[string]string
}{
	"0": {labels: map[string]string{
		kuhbernetes.PodNamespaceLabel: "foons",
		kuhbernetes.PodNameLabel:      "foo",
		containerKindLabel:            "",
	}},
	"1": {labels: map[string]string{
		kuhbernetes.PodNamespaceLabel:     "foons",
		kuhbernetes.PodNameLabel:          "foo",
		kuhbernetes.PodContainerNameLabel: "bar",
	}},
	"2": {labels: map[string]string{
		kuhbernetes.PodNamespaceLabel:     "foons",
		kuhbernetes.PodNameLabel:          "foo",
		kuhbernetes.PodContainerNameLabel: "gnampf",
	}},
}

const cdsock = "/proc/1/root/run/containerd/containerd.sock"

const testref = "docker.io/library/busybox:latest"

var testargs = []string{"/bin/sleep", "30s"}

var nodockerre = regexp.MustCompile(`connect: no such file or directory`)

var _ = Describe("Decorates containerd pod containers", func() {

	var pool *containerdtest.Pool
	var sleepies []*containerdtest.Container

	BeforeEach(func() {
		// In case we're run as root we use a procfs wormhole so we can access
		// the Docker socket even from a test container without mounting it
		// explicitly into the test container.
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		var err error
		pool, err = containerdtest.NewPool(cdsock, "containerd-test")
		Expect(err).NotTo(HaveOccurred())
		for name, config := range names {
			sleepy, err := pool.Run(
				name,
				testref,
				true,
				testargs,
				containerd.WithContainerLabels(config.labels),
			)
			Expect(err).NotTo(HaveOccurred())
			sleepies = append(sleepies, sleepy)
		}
	})

	AfterEach(func() {
		for _, sleepy := range sleepies {
			pool.Purge(sleepy)
		}
	})

	It("decorates k8s pods", func() {
		mw, err := cdwatcher.NewWatcher(cdsock)
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
			id := strings.Split(container.ID, "/")
			if _, ok := names[id[1]]; ok {
				containers = append(containers, container)
			}
		}
		Expect(containers).To(HaveLen(len(names)))

		for _, container := range containers {
			g := container.Group(kuhbernetes.PodGroupType)
			Expect(g).NotTo(BeNil())
			Expect(g.Type).To(Equal(kuhbernetes.PodGroupType))
			Expect(g.Containers).To(ContainElement(container))
			id := strings.Split(container.ID, "/")
			if names[id[1]].labels[kuhbernetes.PodSandboxLabel] != "" {
				Expect(container.Labels).To(HaveKey(kuhbernetes.PodSandboxLabel))
			} else {
				Expect(container.Labels).NotTo(HaveKey(kuhbernetes.PodSandboxLabel))
			}
		}
	})

})
