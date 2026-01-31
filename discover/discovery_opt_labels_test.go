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
	"log/slog"
	"time"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator"
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
	. "github.com/thediveo/success"
)

const testlabelname = "decorator-discovery-label-test"
const testlabelvalue = "test-value"

func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("decorator-label-test"))
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
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))
	})

	const name = "decorator-test-container"

	var sleepy *morbyd.Container

	BeforeAll(func(ctx context.Context) {
		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=discover.labels")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating canary workloads")
		sleepy = Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(name),
			run.WithCommand("/bin/sleep", "120s"),
			run.WithAutoRemove()))
		// Make sure that the newly created container is in running state before we
		// run unit tests which depend on the correct list of alive(!)=running
		// containers.
		Expect(sleepy.PID(ctx)).NotTo(BeZero())
	})

	It("passes discovery labels to decorators", func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		mw, err := moby.New("", nil)
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
