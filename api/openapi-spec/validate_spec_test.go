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

package openapispec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	apitypes "github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var lxknsapispec *openapi3.T
var allns *discover.Result
var pidmap model.PIDMapper

var _ = BeforeSuite(func() {
	var err error
	lxknsapispec, err = openapi3.NewLoader().LoadFromFile("lxkns.yaml")
	Expect(err).To(Succeed())
	Expect(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return lxknsapispec.Validate(ctx)
	}()).To(Succeed(), "lxkns OpenAPI specification is invalid")

	var docksock string
	if os.Geteuid() == 0 {
		docksock = "unix:///proc/1/root/run/docker.sock"
	}

	mw, err := moby.New(docksock, nil)
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := whalefriend.New(ctx, []watcher.Watcher{mw})
	defer cizer.Close()

	<-mw.Ready()

	allns = discover.Namespaces(discover.WithFullDiscovery(), discover.WithContainerizer(cizer))
	pidmap = discover.NewPIDMap(allns)
})

func validate(openapispec *openapi3.T, schemaname string, jsondata []byte) error {
	schemaref, ok := openapispec.Components.Schemas[schemaname]
	if !ok {
		return fmt.Errorf("invalid schema reference %q", schemaname)
	}
	var jsonobj interface{}
	if err := json.Unmarshal(jsondata, &jsonobj); err != nil {
		return err
	}
	return schemaref.Value.VisitJSON(jsonobj)
}

var _ = Describe("lxkns OpenAPI specification", func() {

	It("validates PIDMap", func() {
		j, err := json.Marshal(apitypes.NewPIDMap(apitypes.WithPIDMap(pidmap)))
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "PIDMap", j)).To(Succeed())
	})

	It("validates Process", func() {
		proc := &apitypes.Process{
			PID:           12345,
			PPID:          0,
			ProTaskCommon: model.ProTaskCommon{Name: "foobar"},
		}
		proc.Tasks = append(proc.Tasks, &model.Task{
			TID:           12345,
			ProTaskCommon: proc.ProTaskCommon,
		})
		j, err := json.Marshal(proc)
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "Process", j)).To(Succeed(), string(j))
	})

	It("validates simple ProcessTable", func() {
		proc := &model.Process{PID: 12345, PPID: 0, ProTaskCommon: model.ProTaskCommon{Name: "foobar"}}
		pt := apitypes.NewProcessTable(apitypes.WithProcessTable(
			model.ProcessTable{proc.PID: proc}))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "ProcessTable", j)).To(Succeed(), string(j))
	})

	It("validates (simple) ProcTable", func() {
		j := []byte(`{
			"24566": {
				"pid": 24566,
				"ppid": 3173,
				"name": "bash",
				"cmdline": ["/bin/bash"],
				"namespaces": {},
				"starttime": 745077,
				"cpucgroup": "/user.slice",
				"fridgecgroup": "/fridge.sliced/user",
				"fridgefrozen": true,
				"tasks": [
					{
						"tid": 24566,
						"name": "bash",
						"namespaces": {},
						"starttime": 745077,
						"cpucgroup": "/user.slice",
						"fridgecgroup": "/fridge.sliced/user",
						"fridgefrozen": true
					}
				]
			},
			"2574": {
				"namespaces": {
					"net": 12345678
				},
				"starttime": 51628,
				"cpucgroup": "/user.slice",
				"pid": 2574,
				"ppid": 1,
				"name": "systemd",
				"cmdline": [
					"/lib/systemd/systemd",
					"--user"
				],
				"fridgecgroup": "/outofcontrol",
				"fridgefrozen": false,
				"tasks": [
					{
						"tid": 2574,
						"name": "systemd",
						"namespaces": {},
						"starttime": 51628,
						"cpucgroup": "/user.slice",
						"fridgecgroup": "/outofcontrol",
						"fridgefrozen": false
					}
				]
			}
		}`)
		Expect(validate(lxknsapispec, "ProcessTable", j)).To(Succeed())
	})

	It("validates actual ProcTable", func() {
		pt := apitypes.NewProcessTable(apitypes.WithProcessTable(allns.Processes))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "ProcessTable", j)).To(Succeed(), string(j))
	})

	It("validates a full DiscoveryResult round-trip", func() {
		disco := apitypes.NewDiscoveryResult(apitypes.WithResult(allns))
		j, err := json.Marshal(disco)
		Expect(err).To(Succeed())

		Expect(validate(lxknsapispec, "DiscoveryResult", j)).To(Succeed(), string(j))

		disco2 := apitypes.NewDiscoveryResult()
		Expect(json.Unmarshal(j, disco2)).To(Succeed())
	})

})
