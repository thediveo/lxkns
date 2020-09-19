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
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("PIDMap twin", func() {

	var (
		pidmapjsonA = []byte(`[
			[{"pid": 666,"nsid": 1},{"pid": 1,"nsid": 2}],
			[{"pid": 777,"nsid": 1}]
		]`)
		pidmapjsonB = []byte(`[
			[{"pid": 777,"nsid": 1}]
			[{"pid": 666,"nsid": 1},{"pid": 1,"nsid": 2}],
		]`)
		pidns1 = namespaces.New(species.CLONE_NEWPID, species.NamespaceIDfromInode(1), "")
		pidns2 = namespaces.New(species.CLONE_NEWPID, species.NamespaceIDfromInode(2), "")
		pids   = lxkns.NamespacedPIDs{
			lxkns.NamespacedPID{PID: 666, PIDNS: pidns1},
			lxkns.NamespacedPID{PID: 1, PIDNS: pidns2},
		}
		pids2 = lxkns.NamespacedPIDs{lxkns.NamespacedPID{PID: 777, PIDNS: pidns1}}
		pmap  = lxkns.PIDMap{
			pids[0]:  pids,
			pids[1]:  pids,
			pids2[0]: pids2,
		}
	)

	var (
		allns     *lxkns.DiscoveryResult
		allpidmap lxkns.PIDMap
	)

	BeforeEach(func() {
		discopts := lxkns.NoDiscovery
		discopts.SkipProcs = false
		allns = lxkns.Discover(lxkns.FullDiscovery)
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
			Expect(err).NotTo(HaveOccurred())
			Expect(j).To(Or(MatchJSON(pidmapjsonA), MatchJSON(pidmapjsonB)))
		})

		It("unmarshals PIDMap", func() {
			pmt := NewPIDMap()
			Expect(json.Unmarshal([]byte(`""`), &pmt)).To(HaveOccurred())
			Expect(json.Unmarshal([]byte(`[[]]`), &pmt)).To(MatchError(
				MatchRegexp(`invalid empty list`)))

			Expect(json.Unmarshal(pidmapjsonA, &pmt)).NotTo(HaveOccurred())
			Expect(pmt.PIDMap).To(HaveLen(len(pmap)))
			for _, nspids := range pmt.PIDMap {
				for _, nspid := range nspids {
					Expect(pmap).To(HaveKeyWithValue(nspid, nspids))
				}
			}
		})

		It("survives a full roundtrip without hiccup", func() {
			// Marshal the existing PID map.
			j, err := json.Marshal(NewPIDMap(WithPIDMap(allpidmap)))
			Expect(err).NotTo(HaveOccurred())

			// Unmarshal the JSON soup using the existing PID namespace map.
			pmt2 := NewPIDMap(WithPIDNamespaces(allns.Namespaces[model.PIDNS]))
			Expect(json.Unmarshal(j, &pmt2)).NotTo(HaveOccurred())
			Expect(pmt2.PIDMap).To(HaveLen(len(allpidmap)))
			Expect(pmt2.PIDMap).To(Equal(allpidmap))
		})

	})

})
