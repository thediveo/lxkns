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

// +build linux

package lxkns

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ory/dockertest"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

const sleepyname = "dumb_doormat"

var _ = Describe("Discover containers", func() {

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
	})

	AfterEach(func() {
		Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
	})

	It("finds containers, relates with processes", func() {
		// We cannot discover the initial container process running as root when
		// we're not root too.
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		mw, err := moby.NewWatcher("")
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready() // TODO: should be in whalefriend containerizer

		allns := Discover(WithStandardDiscovery(), WithContainerizer(cizer))

		Expect(allns.Containers).To(ContainElement(
			WithTransform(func(c model.Container) string { return c.Name() }, Equal(sleepyname))))
		var sleepy model.Container
		for _, cntr := range allns.Containers {
			if cntr.Name() == sleepyname {
				sleepy = cntr
				break
			}
		}
		Expect(sleepy).NotTo(BeNil())
		Expect(sleepy.PID()).NotTo(BeZero())
		Expect(allns.Processes).To(HaveKey(sleepy.PID()))
		Expect(sleepy.Process()).NotTo(BeNil())
		Expect(sleepy.Process().Container).To(Equal(sleepy))
	})

})
