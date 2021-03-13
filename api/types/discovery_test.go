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

package types

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	. "github.com/thediveo/lxkns/nstest/gmodel"
)

var _ = Describe("discovery result JSON", func() {

	It("marshals discovery options", func() {
		doh := (*DiscoveryOptions)(&allns.Options)
		j, err := json.Marshal(doh)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(`{
			"skipped-procs": false,
			"skipped-tasks": true,
			"skipped-fds": true,
			"skipped-bindmounts": true,
			"skipped-hierarchy": false,
			"skipped-ownership": false,
			"skipped-freezer": false,
			"scanned-namespace-types": [
			  "time",
			  "mnt",
			  "cgroup",
			  "uts",
			  "ipc",
			  "user",
			  "pid",
			  "net"
			]
		  }`))
	})

	It("unmarshals discovery options", func() {
		opts := &DiscoveryOptions{}
		Expect(json.Unmarshal([]byte(`{
			"scanned-namespace-types": [
			  "foobar"
			]
		  }`),
			opts)).To(MatchError(MatchRegexp("invalid type of namespace")))

		doh := (*DiscoveryOptions)(&allns.Options)
		j, err := json.Marshal(doh)
		Expect(err).To(Succeed())

		opts = &DiscoveryOptions{}
		Expect(json.Unmarshal(j, opts)).To(Succeed())
		Expect(opts).To(Equal(doh))
	})

	It("marshals the discovery options, the namespaces map and process table", func() {
		j, err := json.Marshal(NewDiscoveryResult(WithResult(allns)))
		Expect(err).To(Succeed())

		toplevel := map[string]json.RawMessage{}
		Expect(json.Unmarshal(j, &toplevel)).To(Succeed())

		// "namespaces" must be an object, with keys consisting only of digits.
		Expect(toplevel).To(HaveKey("discovery-options"))
		inner := map[string]json.RawMessage{}
		Expect(json.Unmarshal(toplevel["discovery-options"], &inner)).To(Succeed())
		Expect(inner["scanned-namespace-types"]).To(MatchJSON(`[
			"time",
			"mnt",
			"cgroup",
			"uts",
			"ipc",
			"user",
			"pid",
			"net"
		  ]`))

		// "namespaces" must be an object, with keys consisting only of digits.
		Expect(toplevel).To(HaveKey("namespaces"))
		inner = map[string]json.RawMessage{}
		Expect(json.Unmarshal(toplevel["namespaces"], &inner)).To(Succeed())
		Expect(inner).To(HaveKey(MatchRegexp(`[0-9]+`)))

		// "processes" must be an object, with keys consisting only of digits.
		Expect(toplevel).To(HaveKey("processes"))
		inner = map[string]json.RawMessage{}
		Expect(json.Unmarshal(toplevel["processes"], &inner)).To(Succeed())
		Expect(inner).To(HaveKey(MatchRegexp(`[0-9]+`)))
	})

	It("marshals and unmarshals a discovery results without hiccup", func() {
		j, err := json.Marshal(NewDiscoveryResult(WithResult(allns)))
		Expect(err).To(Succeed())

		dr := NewDiscoveryResult()
		Expect(json.Unmarshal(j, dr)).To(Succeed())
		namespaces := dr.Result().Namespaces
		for idx := model.NamespaceTypeIndex(0); idx < model.NamespaceTypesCount; idx++ {
			Expect(namespaces[idx]).To(HaveLen(len(allns.Namespaces[idx])))
		}
		Expect(dr.Processes()).To(BeSameProcessTable(allns.Processes))

		Expect(dr.Result().Options).To(Equal(allns.Options))
	})

})
