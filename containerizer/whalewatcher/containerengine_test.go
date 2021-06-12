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

package whalewatcher

import (
	"context"

	"github.com/ory/dockertest"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("watching ContainerEngine", func() {

	var pool *dockertest.Pool
	var sleepy *dockertest.Resource

	BeforeEach(func() {
		var err error
		pool, err = dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		sleepy, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       "pompous_pm",
			Cmd:        []string{"/bin/sleep", "30s"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
	})

	It("watches", func() {
		dockerw, err := moby.NewWatcher("")
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cew := New(ctx, []watcher.Watcher{dockerw})
		Expect(cew).NotTo(BeNil())

		<-dockerw.Ready()
		dr := lxkns.DiscoveryResult{}
		cew.Containerize(ctx, &dr)
		Expect(dr.Containers).To(ContainElement(
			WithTransform(func(c model.Container) string { return c.Name() }, Equal("pompous_pm"))))
	})

})
