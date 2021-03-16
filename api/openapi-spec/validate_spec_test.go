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
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/model"
)

var lxknsapispec *openapi3.Swagger
var allns *lxkns.DiscoveryResult
var pidmap lxkns.PIDMap

var _ = BeforeSuite(func() {
	var err error
	lxknsapispec, err = openapi3.NewSwaggerLoader().LoadSwaggerFromFile("lxkns.yaml")
	Expect(err).To(Succeed())
	Expect(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return lxknsapispec.Validate(ctx)
	}()).To(Succeed(), "lxkns OpenAPI specification is invalid")

	allns = lxkns.Discover(lxkns.FullDiscovery)
	pidmap = lxkns.NewPIDMap(allns)
})

func validate(openapispec *openapi3.Swagger, schemaname string, jsondata []byte) error {
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
		j, err := json.Marshal(types.NewPIDMap(types.WithPIDMap(pidmap)))
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "PIDMap", j)).To(Succeed())
	})

	It("validates Process", func() {
		proc := &types.Process{PID: 12345, PPID: 0, Name: "foobar"}
		j, err := json.Marshal(proc)
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "Process", j)).To(Succeed(), string(j))
	})

	It("validates simple ProcessTable", func() {
		proc := &model.Process{PID: 12345, PPID: 0, Name: "foobar"}
		pt := types.NewProcessTable(types.WithProcessTable(
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
				"fridgefrozen": true
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
				"fridgefrozen": false
			}
		}`)
		Expect(validate(lxknsapispec, "ProcessTable", j)).To(Succeed())
	})

	It("validates actual ProcTable", func() {
		pt := types.NewProcessTable(types.WithProcessTable(allns.Processes))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())
		Expect(validate(lxknsapispec, "ProcessTable", j)).To(Succeed(), string(j))
	})

	It("validates a full DiscoveryResult round-trip", func() {
		disco := types.NewDiscoveryResult(types.WithResult(allns))
		j, err := json.Marshal(disco)
		Expect(err).To(Succeed())

		Expect(validate(lxknsapispec, "DiscoveryResult", j)).To(Succeed(), string(j))

		disco2 := types.NewDiscoveryResult()
		Expect(json.Unmarshal(j, disco2)).To(Succeed())
	})

})
