//go:build podman

// Copyright 2022 Harald Albrecht.
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

package podman

import (
	"context"
	"os"
	"time"

	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/pods"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/containers/podman/v3/pkg/rootless"
	"github.com/containers/podman/v3/pkg/specgen"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/test/matcher"
	"github.com/thediveo/sealwatcher"
	"github.com/thediveo/sealwatcher/test"
	"github.com/thediveo/whalewatcher"
	"github.com/thediveo/whalewatcher/watcher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/whalewatcher/test/matcher"
)

const dimLizzyPod = "DimLizzyPod"

var (
	dimLizzy = test.NewContainerDescription{
		Name:   "dimm_lizzy",
		Status: test.Running,
	}
)

var _ = Describe("Decorates Podman pods", Serial, func() {

	// Ensure to run the goroutine leak test *last* after all (defered)
	// clean-ups.
	BeforeEach(func() {
		if os.Getuid() != 0 || rootless.IsRootless() /* work around botched podman code base */ {
			Skip("needs root")
		}

		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("decorated Podman pods", slowSpec, func(ctx context.Context) {
		By("watching seals")
		pw, err := sealwatcher.New("unix:///run/podman/podman.sock", nil)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{pw})
		defer cizer.Close()

		<-pw.Ready()

		By("cleaning up any leftover test pod")
		podconn, err := bindings.NewConnection(ctx, "unix:///run/podman/podman.sock")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			conn, _ := bindings.GetClient(podconn)
			conn.Client.CloseIdleConnections()
		}()

		force := true
		_, _ = pods.Remove(podconn, dimLizzyPod, &pods.RemoveOptions{Force: &force})

		By("creating a pod")
		_, err = pods.CreatePodFromSpec(podconn, &entities.PodSpec{
			PodSpecGen: specgen.PodSpecGenerator{
				PodBasicConfig: specgen.PodBasicConfig{
					Name: dimLizzyPod,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_, _ = pods.Remove(podconn, dimLizzyPod, &pods.RemoveOptions{Force: &force})
		}()
		id := test.NewContainer(podconn, dimLizzy, test.OfPod(dimLizzyPod))

		By("finding the pod")
		Eventually(func() *whalewatcher.Container {
			return pw.Portfolio().Project("").Container(dimLizzy.Name)
		}).WithTimeout(5 * time.Second).ShouldNot(BeNil())

		allcontainers := cizer.Containers(ctx, model.NewProcessTable(false), nil)
		var canary *model.Container
		Expect(allcontainers).To(ContainElement(And(
			HaveID(id),
			matcher.WithType(sealwatcher.Type),
		), &canary))

		By("decorating the pod's container")
		Decorate([]*model.ContainerEngine{canary.Engine}, nil)

		Expect(canary).Should(And(
			HaveID(id),
			matcher.BeInAGroup(
				HaveName(dimLizzyPod),
				HaveField("Type", PodGroupType),
				HaveField("Flavor", PodGroupType),
			),
		))

		By("done")
	})

})
