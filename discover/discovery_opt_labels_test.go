// Copyright 2022 Harald Albrecht.
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

package discover

import (
	"context"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/go-plugger"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/noleak"
)

const testlabelname = "decorator-discovery-label-test"
const testlabelvalue = "test-value"

func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Name:  "decorator-label-test",
		Group: decorator.PluginGroup,
		Symbols: []plugger.Symbol{
			decorator.Decorate(Decorate),
		},
	})
}

func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	for _, engine := range engines {
		for _, c := range engine.Containers {
			c.Labels[testlabelname] = labels[testlabelname]
		}
	}
}

var _ = Describe("decorator discovery labels", Ordered, func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
		})
	})

	const name = "decorator-test-container"

	var pool *dockertest.Pool
	var sleepy *dockertest.Resource
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
		_ = pool.RemoveContainerByName(name)
		Eventually(func() error {
			_, err := pool.Client.InspectContainer(name)
			return err
		}, "5s", "100ms").Should(HaveOccurred())
		sleepy, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       name,
			Cmd:        []string{"/bin/sleep", "120s"},
			Labels:     map[string]string{},
		})
		// Skip test in case Docker is not accessible.
		if err != nil && noDockerRE.MatchString(err.Error()) {
			Skip("Docker not available")
		}
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() bool {
			c, err := pool.Client.InspectContainer(sleepy.Container.ID)
			Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
			return c.State.Running
		}, "5s", "100ms").Should(BeTrue(), "container %s", sleepy.Container.Name[1:])

		DeferCleanup(func() {
			Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
			pool.Client.HTTPClient.CloseIdleConnections()
		})
	})

	It("passes discovery labels to decorators", func() {
		mw, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		allns := Namespaces(
			WithStandardDiscovery(),
			WithLabel(testlabelname, testlabelvalue),
			WithContainerizer(cizer))
		Expect(allns.Containers).To(ContainElement(
			HaveValue(HaveField("Labels", HaveKeyWithValue(testlabelname, testlabelvalue)))))

		allns = Namespaces(
			WithStandardDiscovery(),
			WithContainerizer(cizer))
		Expect(allns.Containers).NotTo(ContainElement(
			HaveValue(HaveField("Labels", HaveKeyWithValue(testlabelname, testlabelvalue)))))

	})

})
