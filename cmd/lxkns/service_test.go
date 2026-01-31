// Copyright 2020 Harald Albrecht.
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

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

// workload test container "random" name
var sleepyname = "empty_eno" + strconv.FormatInt(GinkgoRandomSeed(), 10)

// where to reach the API after starting the service on a dynamic port.
var baseurl string

var _ = Describe("serves API endpoints", Ordered, func() {

	BeforeAll(func(ctx context.Context) {
		docksock := ""
		if os.Geteuid() == 0 {
			docksock = "unix:///proc/self/root/run/docker.sock"
		}

		// "hardcore" check after all has been said and done
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(7 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		By("creating a new Docker session for testing")
		sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("lxkns.test=cmd.lxkns")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating a canary workload")
		sleepy := Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(sleepyname),
			run.WithCommand("/bin/sleep", "120s"),
			run.WithAutoRemove()))
		// Make sure that the newly created container is in running state before we
		// run unit tests which depend on the correct list of alive(!)=running
		// containers.
		Expect(sleepy.PID(ctx)).NotTo(BeZero())

		By("starting a workload watcher")
		moby, err := moby.New(docksock, nil)
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
			moby.Close()
		})

		By("starting a containerizer")
		cizer := whalefriend.New(ctx, []watcher.Watcher{moby})
		DeferCleanup(func() { cizer.Close() })
		Eventually(moby.Ready, "5s").Should(BeClosed())

		By("starting the service")
		oldSlogger := slog.Default()
		DeferCleanup(func() {
			slog.SetDefault(oldSlogger)
		})
		slog.SetDefault(slog.New(
			slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
				Level: slog.LevelError + 4,
			})))
		serveraddr, err := startServer("127.0.0.1:0", cizer)
		Expect(err).NotTo(HaveOccurred())
		baseurl = "http://" + serveraddr.String() + "/api/"
		DeferCleanup(func() {
			stopServer(5 * time.Second)
		})
	})

	BeforeEach(func() {
		goods := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(
					goods,
					// as we have no direct control over the workerpool worker
					// goroutines we ignore them in this specific case.
					IgnoringCreator("github.com/gammazero/workerpool.(*WorkerPool).dispatch")))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("cannot find non-existing API", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		defer clnt.CloseIdleConnections()
		resp, err := clnt.Get(baseurl + "foobar")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("discovers namespaces", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		defer clnt.CloseIdleConnections()
		resp, err := clnt.Get(baseurl + "namespaces")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		allns := types.NewDiscoveryResult()
		Expect(json.NewDecoder(resp.Body).Decode(allns)).To(Succeed())
		for idx, nsmap := range allns.Result().Namespaces {
			if model.NamespaceTypeIndex(idx) != model.TimeNS {
				Expect(nsmap).NotTo(BeEmpty())
			}
		}
		Expect(allns.Result().Processes).NotTo(BeEmpty())
		Expect(allns.ContainerModel.Containers.Containers).To(ContainElement(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(sleepyname),
			})),
		))
	})

	It("discovers processes", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		defer clnt.CloseIdleConnections()
		resp, err := clnt.Get(baseurl + "processes")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		procs := types.NewProcessTable()
		Expect(json.NewDecoder(resp.Body).Decode(&procs)).To(Succeed())
		Expect(procs.ProcessTable).NotTo(BeEmpty())
	})

	It("discovers pid mapping", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		defer clnt.CloseIdleConnections()
		resp, err := clnt.Get(baseurl + "pidmap")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		pidmap := types.NewPIDMap()
		Expect(json.NewDecoder(resp.Body).Decode(&pidmap)).To(Succeed())
		Expect(pidmap.PIDMap).NotTo(BeEmpty())
	})

})
