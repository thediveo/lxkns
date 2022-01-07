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
	"net/http"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

var sleepyname = "empty_eno" + strconv.FormatInt(GinkgoRandomSeed(), 10)

var baseurl string

var pool *dockertest.Pool
var sleepy *dockertest.Resource
var ctx context.Context
var cancel context.CancelFunc

var _ = BeforeSuite(func() {
	docksock := ""
	if os.Geteuid() == 0 {
		docksock = "unix:///proc/1/root/run/docker.sock"
	}

	var err error
	pool, err = dockertest.NewPool(docksock)
	Expect(err).NotTo(HaveOccurred())
	_ = pool.RemoveContainerByName(sleepyname)
	Eventually(func() error {
		_, err := pool.Client.InspectContainer(sleepyname)
		return err
	}, "5s").Should(HaveOccurred())
	sleepy, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "busybox",
		Tag:        "latest",
		Name:       sleepyname,
		Cmd:        []string{"/bin/sleep", "120s"},
	})
	Expect(err).NotTo(HaveOccurred(), "container %s", sleepyname)
	Eventually(func() bool {
		c, err := pool.Client.InspectContainer(sleepy.Container.ID)
		Expect(err).NotTo(HaveOccurred(), "container %s", sleepy.Container.Name[1:])
		return c.State.Running
	}, "5s", "100ms").Should(BeTrue(), "container %s", sleepy.Container.Name[1:])

	moby, err := moby.New(docksock, nil)
	Expect(err).NotTo(HaveOccurred())
	ctx, cancel = context.WithCancel(context.Background())
	cizer := whalefriend.New(ctx, []watcher.Watcher{moby})
	Eventually(moby.Ready, "5s").Should(BeClosed())

	log.SetLevel(log.FatalLevel)
	serveraddr, err := startServer("127.0.0.1:0", cizer)
	Expect(err).To(Succeed())
	baseurl = "http://" + serveraddr.String() + "/api/"
})

var _ = AfterSuite(func() {
	if sleepy != nil {
		Expect(pool.Purge(sleepy)).NotTo(HaveOccurred())
	}
	stopServer(5 * time.Second)
	if cancel != nil {
		cancel()
	}
})

var _ = Describe("serves API endpoints", func() {

	It("cannot find non-existing API", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		resp, err := clnt.Get(baseurl + "foobar")
		Expect(err).To(Succeed())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("discovers namespaces", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		resp, err := clnt.Get(baseurl + "namespaces")
		Expect(err).To(Succeed())
		defer resp.Body.Close()
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
		resp, err := clnt.Get(baseurl + "processes")
		Expect(err).To(Succeed())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		procs := types.NewProcessTable()
		Expect(json.NewDecoder(resp.Body).Decode(&procs)).To(Succeed())
		Expect(procs.ProcessTable).NotTo(BeEmpty())
	})

	It("discovers pid mapping", func() {
		clnt := &http.Client{Timeout: 10 * time.Second}
		resp, err := clnt.Get(baseurl + "pidmap")
		Expect(err).To(Succeed())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		pidmap := types.NewPIDMap()
		Expect(json.NewDecoder(resp.Body).Decode(&pidmap)).To(Succeed())
		Expect(pidmap.PIDMap).NotTo(BeEmpty())
	})

})
