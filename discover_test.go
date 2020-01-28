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

package lxkns

import (
	"encoding/json"
	"os/exec"

	t "github.com/thediveo/lxkns/nstypes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type lsnsentry struct {
	NS      t.NamespaceID `json:"ns"`
	Type    string        `json:"type"`
	NProcs  int           `json:"nprocs"`
	PID     PIDType       `json:"pid"`
	User    string        `json:"user"`
	Command string        `json:"command"`
}

type lsnsdata struct {
	Namespaces []lsnsentry `json:"namespaces"`
}

func lsns(opts ...string) []lsnsentry {
	out, err := exec.Command(
		"lsns",
		append([]string{"--json"}, opts...)...).Output()
	Expect(err).NotTo(HaveOccurred())
	var res lsnsdata
	err = json.Unmarshal(out, &res)
	Expect(err).NotTo(HaveOccurred())
	Expect(res.Namespaces).NotTo(BeEmpty())
	return res.Namespaces
}

var _ = Describe("Discover", func() {

	It("sees the namespaces lsns sees", func() {
		allns := Discover(FullDiscovery)
		for _, ns := range lsns() {
			nsidx := TypeIndex(t.NameToType(ns.Type))
			dns := allns.Namespaces[nsidx][ns.NS]
			Expect(dns).NotTo(BeZero())
			Expect(dns.LeaderPIDs()).To(ContainElement(PIDType(ns.PID)))
		}
	})

})
