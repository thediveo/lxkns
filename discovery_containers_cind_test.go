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

package lxkns

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	cdengine "github.com/thediveo/whalewatcher/engineclient/containerd"
	mobyengine "github.com/thediveo/whalewatcher/engineclient/moby"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

var _ = Describe("Discover containers in containers", func() {

	It("translates container PIDs", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		pt := model.NewProcessTable(false)
		var mobypid model.PIDType
		for _, proc := range pt {
			if proc.Name == "dockerd" {
				mobypid = proc.PID
				break
			}
		}
		Expect(mobypid).NotTo(BeZero(), "dockerd not found")

		mw, err := moby.New("", nil, mobyengine.WithPID(int(mobypid)))
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		Eventually(mw.Ready()).Should(BeClosed(), "dockerd watcher failed to synchronize")

		allns := Discover(WithStandardDiscovery(), WithContainerizer(cizer))
		Expect(allns.Containers).To(ContainElement(
			WithTransform(func(c *model.Container) string { return c.Name }, Equal("containerd-in-docker"))))

		var cind *model.Container
		for _, c := range allns.Containers {
			if c.Name == "containerd-in-docker" {
				cind = c
				break
			}
		}
		enginepid := cind.PID
		Expect(enginepid).NotTo(BeZero(), "missing/invalid container %q with zero PID", cind.Name)
		cancel()

		mw, err = moby.New("", nil, mobyengine.WithPID(int(mobypid)))
		Expect(err).NotTo(HaveOccurred())

		cdw, err := containerd.New(
			fmt.Sprintf("///proc/%d/root/run/containerd/containerd.sock", enginepid), nil, cdengine.WithPID(int(enginepid)))
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		cizer = whalefriend.New(ctx, []watcher.Watcher{mw, cdw})
		defer cizer.Close()

		Eventually(mw.Ready()).Should(BeClosed(), "dockerd watcher failed to synchronize")
		Eventually(cdw.Ready()).Should(BeClosed(), "containerd watcher failed to synchronize")

		allns = Discover(WithStandardDiscovery(), WithContainerizer(cizer))
		Expect(allns.Containers).To(ContainElement(
			WithTransform(func(c *model.Container) string { return c.Engine.Type }, Equal(cdengine.Type))))
		var sleepy *model.Container
		for _, c := range allns.Containers {
			if c.Engine.Type == cdengine.Type {
				Expect(sleepy).To(BeZero())
				sleepy = c
			}
		}
		Expect(sleepy.Process).NotTo(BeNil())
		Expect(sleepy.Process.Cmdline).To(ConsistOf("sleep", ContainSubstring("1000")))

	})

})
