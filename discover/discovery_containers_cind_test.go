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
	"log/slog"
	"os"
	"syscall"
	"time"

	cd "github.com/containerd/containerd"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/build"
	"github.com/thediveo/morbyd/exec"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/morbyd/timestamper"
	cdengine "github.com/thediveo/whalewatcher/engineclient/containerd"
	"github.com/thediveo/whalewatcher/engineclient/cri/test/img"
	mobyengine "github.com/thediveo/whalewatcher/engineclient/moby"
	"github.com/thediveo/whalewatcher/test"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	imgName  = "thediveo/kindisch-lxkns-containerd"
	cindName = "lxkns-cind" // name of Docker container with containerd
)

var _ = Describe("Discovering containers in containers", Serial, func() {

	var sess *morbyd.Session
	var providerCntr *morbyd.Container

	// Ensure to run the goroutine leak test *last* after all (deferred)
	// clean-ups.
	BeforeEach(slowSpec, func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
			return
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		By("creating a new Docker session for testing")
		sess = Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=discover")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		// For details, please see:
		// https://github.com/thediveo/whalewatcher/blob/cca7f5676b3f63b0e2d6311a60ca3da2fd07ead7/engineclient/containerd/containerd_test.go#L115
		By("spinning up a Docker container with stand-alone containerd, courtesy of the KinD k8s sig")
		Expect(sess.BuildImage(ctx, "./test/_kindisch",
			build.WithTag(imgName),
			build.WithBuildArg("KINDEST_BASE_TAG="+test.KindestBaseImageTag),
			build.WithOutput(timestamper.New(GinkgoWriter)))).
			Error().NotTo(HaveOccurred())
		providerCntr = Successful(sess.Run(ctx, img.Name,
			run.WithName(cindName),
			run.WithAutoRemove(),
			run.WithPrivileged(),
			run.WithSecurityOpt("label=disable"),
			run.WithCgroupnsMode("private"),
			run.WithVolume("/var"),
			run.WithVolume("/dev/mapper:/dev/mapper"),
			run.WithVolume("/lib/modules:/lib/modules:ro"),
			run.WithTmpfs("/tmp"),
			run.WithTmpfs("/run"),
			run.WithDevice("/dev/fuse"),
			run.WithCombinedOutput(timestamper.New(GinkgoWriter))))
		DeferCleanup(func(ctx context.Context) {
			By("removing the test container")
			providerCntr.Kill(ctx)
		})

		By("waiting for containerized containerd to become responsive")
		pid := Successful(providerCntr.PID(ctx))
		// apipath must not include absolute symbolic links, but already be
		// properly resolved.
		endpointPath := fmt.Sprintf("/proc/%d/root%s",
			pid, "/run/containerd/containerd.sock")
		var cdclient *cd.Client
		Eventually(func() error {
			var err error
			cdclient, err = cd.New(endpointPath,
				cd.WithTimeout(5*time.Second))
			return err
		}).Within(30*time.Second).ProbeEvery(1*time.Second).
			Should(Succeed(), "containerd API never became responsive")
		_ = cdclient.Close() // not needed anymore, will create fresh ones over and over again

		By("creating a dummy containerd workload that runs detached")
		cmd := Successful(providerCntr.Exec(ctx,
			exec.Command("ctr",
				"image", "pull",
				"docker.io/library/busybox:latest"),
			exec.WithCombinedOutput(timestamper.New(GinkgoWriter))))
		Expect(cmd.Wait(ctx)).To(BeZero())
		cmd = Successful(providerCntr.Exec(ctx,
			exec.Command("ctr",
				"run",
				"--detach",
				"--label", "name=sleepy",
				"docker.io/library/busybox:latest",
				"sleepy",
				"/bin/sh", "-c", "while true; do sleep 1; echo -n .; done"),
			exec.WithCombinedOutput(timestamper.New(GinkgoWriter))))
		Expect(cmd.Wait(ctx)).To(BeZero())
	})

	It("translates container-in-container PIDs", slowSpec, func(ctx context.Context) {
		By("finding the right Docker daemon PID (too many mobys these days *scnr*)")
		// use /run/docker.sock for consistency, avoid symlinks later!
		dockerSockIno := Successful(
			os.Stat("/run/docker.sock")).Sys().(*syscall.Stat_t).Ino

		mobyprocs := model.NewProcessTable(false).ByName("dockerd")
		var mobyproc *model.Process
		Expect(mobyprocs).To(ContainElement(
			WithTransform(func(proc *model.Process) uint64 {
				return Successful(
					os.Stat(fmt.Sprintf("/proc/%d/root/run/docker.sock", proc.PID))).
					Sys().(*syscall.Stat_t).Ino
			},
				Equal(dockerSockIno)), &mobyproc))
		mobypid := mobyproc.PID
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

		By("watching the Dockerized containerd")
		// we're lazy here and just use the Docker container's PID instead of
		// the Dockerized containerd's PID, but that's close enough here.
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
			Not(HaveField("Cmdline", ConsistOf("sleep", ContainSubstring("1"))))))

		By("looking for the sleepy container, now with a PID mapper")
		allns = Namespaces(WithStandardDiscovery(), WithContainerizer(cizer), WithPIDMapper())
		Expect(allns.PIDMap).NotTo(BeNil())
		containerds = allns.Containers.WithEngineType(cdengine.Type)
		Expect(containerds).To(HaveLen(1))
		sleepy = containerds[0]
		Expect(sleepy.Labels).To(HaveKeyWithValue("name", "sleepy"))
		Expect(sleepy.Process).NotTo(BeNil())
		Expect(sleepy.Process.Cmdline).To(ConsistOf("/bin/sh", "-c", ContainSubstring("sleep 1")))
		Expect(sleepy.PID).To(Equal(sleepy.Process.PID))
	})

})
