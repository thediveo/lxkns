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

package discover

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	cdengine "github.com/thediveo/whalewatcher/engineclient/containerd"
	mobyengine "github.com/thediveo/whalewatcher/engineclient/moby"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

const cindName = "containerd-in-docker" // name of Docker container with containerd

var _ = Describe("Discovering containers in containers", Serial, func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(slowSpec, func(_ context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
			return
		}

		By("setting things up, hopefully not upsetting them")
		out, err := exec.Command("./test/cind/setup.sh").CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "with output:", out)
		DeferCleanup(slowSpec, func(_ context.Context) {
			By("tearing things down")
			out, err := exec.Command("./test/cind/teardown.sh").CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "with output:", out)
		})

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("translates container-in-container PIDs", slowSpec, func(ctx context.Context) {
		By("finding the Docker daemon PID")
		mobyproc := model.NewProcessTable(false).ByName("dockerd")
		Expect(mobyproc).To(HaveLen(1))
		mobypid := mobyproc[0].PID
		Expect(mobypid).NotTo(BeZero())

		By("watching the Docker daemon with a known PID")
		mw, err := moby.New("", nil, mobyengine.WithPID(int(mobypid)))
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()
		Eventually(mw.Ready()).Should(BeClosed(), "dockerd watcher failed to synchronize")

		By("finding the containerd-in-docker container")
		allns := Namespaces(WithStandardDiscovery(), WithContainerizer(cizer))
		cind := allns.Containers.FirstWithName(cindName)
		Expect(cind).NotTo(BeNil())
		enginepid := cind.PID
		Expect(enginepid).NotTo(BeZero(), "missing/invalid container %q with zero PID", cind.Name)
		cancel()

		By("watching both the Docker daemon and containerd")
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

		By("without a PIDMapper looking for the sleepy container with the sleep process inside the containerd-in-docker container")
		allns = Namespaces(WithStandardDiscovery(), WithContainerizer(cizer))
		Expect(allns.PIDMap).To(BeNil()) //!!!ensure we don't have any mapping available
		containerds := allns.Containers.WithEngineType(cdengine.Type)
		Expect(containerds).To(HaveLen(1))
		sleepy := containerds[0]
		Expect(sleepy.Labels).To(HaveKeyWithValue("name", "sleepy"))
		Expect(sleepy.Process).To(Or(
			BeNil(),
			Not(HaveField("Cmdline", ConsistOf("sleep", ContainSubstring("1000"))))))

		By("looking for the sleepy container, now with a PID mapper")
		allns = Namespaces(WithStandardDiscovery(), WithContainerizer(cizer), WithPIDMapper())
		Expect(allns.PIDMap).NotTo(BeNil())
		containerds = allns.Containers.WithEngineType(cdengine.Type)
		Expect(containerds).To(HaveLen(1))
		sleepy = containerds[0]
		Expect(sleepy.Labels).To(HaveKeyWithValue("name", "sleepy"))
		Expect(sleepy.Process).NotTo(BeNil())
		Expect(sleepy.Process.Cmdline).To(ConsistOf("/bin/sh", "-c", ContainSubstring("sleep 1000")))
		Expect(sleepy.PID).To(Equal(sleepy.Process.PID))
	})

})
