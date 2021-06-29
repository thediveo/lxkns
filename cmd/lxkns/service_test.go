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
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
)

var baseurl string

var _ = BeforeSuite(func() {
	log.SetLevel(log.FatalLevel)
	serveraddr, err := startServer("127.0.0.1:0")
	Expect(err).To(Succeed())
	baseurl = "http://" + serveraddr.String() + "/api/"
})

var _ = AfterSuite(func() {
	stopServer(5 * time.Second)
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
