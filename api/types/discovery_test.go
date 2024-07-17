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

	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/lxkns/nstest/gmodel"
)

var _ = Describe("discovery result JSON", func() {

	It("marshals discovery options", func() {
		doh := (*DiscoveryOptions)(&allns.Options)
		j, err := json.Marshal(doh)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(`{
			"from-procs": true,
			"from-tasks": false,
			"from-fds": false,
			"from-bindmounts": false,
			"with-hierarchy": true,
			"with-ownership": true,
			"with-freezer": true,
			"with-mounts": true,
			"with-socket-processes": false,
			"with-affinity-scheduling": false,
			"labels": {},
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
		// Slightly convoluted test for correct options sans the non-marshalled
		// ones...
		jopts, err := json.Marshal(opts)
		Expect(err).NotTo(HaveOccurred())
		jdoh, err := json.Marshal(doh)
		Expect(err).NotTo(HaveOccurred())
		Expect(jopts).To(MatchJSON(jdoh))
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

		// "mounts" must be an object, with keys consisting only of digits.
		Expect(toplevel).To(HaveKey("mounts"))
		inner = map[string]json.RawMessage{}
		Expect(json.Unmarshal(toplevel["mounts"], &inner)).To(Succeed())
		Expect(inner).To(HaveKey(MatchRegexp(`[0-9]+`)))
	})

	It("marshals and unmarshals a discovery result without hiccup", func() {
		j, err := json.Marshal(NewDiscoveryResult(WithResult(allns)))
		Expect(err).To(Succeed())

		dr := NewDiscoveryResult()
		Expect(json.Unmarshal(j, dr)).To(Succeed())
		namespaces := dr.Result().Namespaces
		for idx := model.NamespaceTypeIndex(0); idx < model.NamespaceTypesCount; idx++ {
			Expect(namespaces[idx]).To(HaveLen(len(allns.Namespaces[idx])))
		}
		Expect(dr.Processes()).To(BeSameProcessTable(allns.Processes))

		Expect(len(dr.Mounts())).To(Equal(len(allns.Mounts)))
		for mntnsid, m := range dr.Mounts() {
			Expect(len(m)).To(Equal(len(allns.Mounts[mntnsid])))
		}

		// Slightly convoluted test for correct options sans the non-marshalled
		// ones...
		opts, err := json.Marshal(dr.Result().Options)
		Expect(err).NotTo(HaveOccurred())
		allnsopts, err := json.Marshal(allns.Options)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts).To(MatchJSON(allnsopts))

		// Coarse check that the numbers of containers, engines, and groups
		// match.
		Expect(dr.Result().Containers).To(HaveLen(len(allns.Containers)))
		drcm := NewContainerModel(dr.Result().Containers)
		allnscm := NewContainerModel(allns.Containers)
		Expect(drcm.Containers.Containers).To(HaveLen(len(allnscm.Containers.Containers)))
		Expect(drcm.ContainerEngines.engineRefIDs).To(HaveLen(len(allnscm.ContainerEngines.engineRefIDs)))
		Expect(drcm.Groups.groupRefIDs).To(HaveLen(len(allnscm.Groups.groupRefIDs)))
	})

})
