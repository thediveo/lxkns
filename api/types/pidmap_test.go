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
	"fmt"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/internal/pidmap"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("PIDMap twin", func() {

	var (
		pidmapjson = `[
			[{"pid": 666,"nsid": 1},{"pid": 1,"nsid": 2}],
			[{"pid": 777,"nsid": 1}]
		]`
		rootpidns   = namespaces.New(species.CLONE_NEWPID, species.NamespaceIDfromInode(1), "")
		pidns2      = namespaces.New(species.CLONE_NEWPID, species.NamespaceIDfromInode(2), "")
		proc666pids = model.NamespacedPIDs{
			model.NamespacedPID{PID: 666, PIDNS: rootpidns},
			model.NamespacedPID{PID: 1, PIDNS: pidns2},
		}
		proc777pids = model.NamespacedPIDs{model.NamespacedPID{PID: 777, PIDNS: rootpidns}}
		pmap        = pidmap.PIDMap{
			proc666pids[0]: proc666pids,
			proc666pids[1]: proc666pids,
			proc777pids[0]: proc777pids,
		}
	)

	var (
		allns     *lxkns.DiscoveryResult
		allpidmap model.PIDMapper
	)

	BeforeEach(func() {
		allns = lxkns.Discover(lxkns.WithStandardDiscovery())
		allpidmap = lxkns.NewPIDMap(allns)
	})

	Describe("JSON un/marshaller", func() {

		It("creates PIDMap twins with options", func() {
			pmt := NewPIDMap()
			Expect(pmt.PIDMap).NotTo(BeNil())
			Expect(pmt.PIDns).NotTo(BeNil())

			pidmap := lxkns.NewPIDMap(allns)
			pmt = NewPIDMap(WithPIDMap(pidmap))
			Expect(pmt.PIDMap).To(Equal(pidmap))

			pmt = NewPIDMap(WithPIDNamespaces(allns.Namespaces[model.PIDNS]))
			Expect(pmt.PIDns).To(Equal(allns.Namespaces[model.PIDNS]))
		})

		It("marshals PIDMap", func() {
			pmt := NewPIDMap(WithPIDMap(pmap))
			j, err := json.Marshal(pmt)
			Expect(err).To(Succeed())
			obj := []namespacedPIDs{}
			Expect(json.Unmarshal(j, &obj)).To(Succeed())
			Expect(obj).To(HaveLen(2))
			Expect(obj).To(ConsistOf(
				ConsistOf(
					namespacedPID{PID: 1, NamespaceID: 2},
					namespacedPID{PID: 666, NamespaceID: 1}),
				ConsistOf(namespacedPID{PID: 777, NamespaceID: 1})))
		})

		It("unmarshals PIDMap", func() {
			pmt := NewPIDMap()
			Expect(json.Unmarshal([]byte(`""`), &pmt)).To(HaveOccurred())
			Expect(json.Unmarshal([]byte(`[[]]`), &pmt)).To(MatchError(
				MatchRegexp(`invalid empty list`)))

			Expect(json.Unmarshal([]byte(pidmapjson), &pmt)).To(Succeed())
			Expect(pmt.PIDMap).To(HaveLen(len(pmap)))
			for _, nspids := range pmt.PIDMap.(pidmap.PIDMap) {
				for _, nspid := range nspids {
					Expect(pmap).To(HaveKeyWithValue(nspid, nspids))
				}
			}
		})

		It("survives a full roundtrip without hiccup", func() {
			// Marshal the existing PID map.
			j, err := json.Marshal(NewPIDMap(WithPIDMap(allpidmap)))
			Expect(err).To(Succeed())

			// Unmarshal the JSON soup using the existing PID namespace map.
			pmt2 := NewPIDMap(WithPIDNamespaces(allns.Namespaces[model.PIDNS]))
			dumponerror := func() string {
				s := "un/marshalling PID map size error\n"
				s += fmt.Sprintf("expected/unmarshalled: len %d\n%s\n", len(pmt2.PIDMap.(pidmap.PIDMap)), sortedpidmap(pmt2.PIDMap.(pidmap.PIDMap)))
				s += fmt.Sprintf("actual/marshalled: len %d\n%s", len(allpidmap.(pidmap.PIDMap)), sortedpidmap(allpidmap.(pidmap.PIDMap)))
				return s
			}
			Expect(json.Unmarshal(j, &pmt2)).To(Succeed())
			Expect(len(pmt2.PIDMap.(pidmap.PIDMap))).To(Equal(len(allpidmap.(pidmap.PIDMap))), dumponerror)
			Expect(pmt2.PIDMap).To(Equal(allpidmap), dumponerror)
		})

	})

})

func sortedpidmap(pm pidmap.PIDMap) string {
	s := []string{}
	for nspid, nspids := range pm {
		l := []string{}
		for _, nspid := range nspids {
			l = append(l, fmt.Sprintf(
				"(%d, pid:[%d])", nspid.PID, nspid.PIDNS.ID().Ino))
		}
		s = append(s, fmt.Sprintf(
			"\t%6d pid:[%d]: [ %s ]",
			nspid.PID, nspid.PIDNS.ID().Ino,
			strings.Join(l, ", ")))
	}
	sort.Strings(s)
	return strings.Join(s, "\n")
}
