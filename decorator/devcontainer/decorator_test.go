// Copyright 2026 Harald Albrecht.
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

package devcontainer

import (
	"context"
	"log/slog"
	"time"

	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/matcher"
	. "github.com/thediveo/success"
)

var _ = Describe("decorating devcontainers and codespaces", func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("decorates plain devcontainers", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("test=decorator.devcontainer")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("sordid_sufferer"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerLocalfolderLabelName+"="),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("baseless_boris"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerLocalfolderLabelName+"=/"),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("devious_devcontainer"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerLocalfolderLabelName+"="+"/home/foo/githog/myproject"),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		By("watcher whales")
		mw, err := moby.New("", nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		By("finding a Docker engine")
		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())

		var canaries []*model.Container
		Expect(allcontainers).To(ContainElement(WithType(moby.Type), &canaries))

		composer.Decorate([]*model.ContainerEngine{canaries[0].Engine}, nil)
		Decorate([]*model.ContainerEngine{canaries[0].Engine}, nil)

		Expect(canaries[0].Engine.Containers).To(ContainElement(
			HaveLabel(DevContainerNameLabelName, "myproject")))
	})

	It("decorates codespace containers", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("test=decorator.devcontainer")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("stupid_stumper"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerLocalfolderLabelName+"="),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("pretty_prattle"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerMetadataLabelName+
				`=[{},{"containerEnv":{"CODESPACES":"???"}},{}]`),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		Expect(sess.Run(ctx, "busybox",
			run.WithAutoRemove(),
			run.WithName("kurious_kaleidoscope"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithLabel(DevcontainerMetadataLabelName+
				`=[{},{"containerEnv":{"CODESPACES":"true","RepositoryName":"furious furuncle"}},{}]`),
			run.WithCombinedOutput(GinkgoWriter),
		)).Error().NotTo(HaveOccurred())

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		By("watcher whales")
		mw, err := moby.New("", nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
		defer cizer.Close()

		<-mw.Ready()

		By("finding a Docker engine")
		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		Expect(allcontainers).NotTo(BeEmpty())

		var canaries []*model.Container
		Expect(allcontainers).To(ContainElement(WithType(moby.Type), &canaries))

		composer.Decorate([]*model.ContainerEngine{canaries[0].Engine}, nil)
		Decorate([]*model.ContainerEngine{canaries[0].Engine}, nil)

		Expect(canaries[0].Engine.Containers).To(ContainElement(
			HaveLabel(CodespaceNameLabelName, "furious furuncle")))
	})

})
