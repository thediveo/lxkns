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
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Discover", func() {

	It("discovers nothing unless told so", func() {
		allns := Discover(NoDiscovery)
		for _, nsmap := range allns.Namespaces {
			Expect(nsmap).To(HaveLen(0))
		}
	})

	It("sorts namespace maps", func() {
		nsmap := NamespaceMap{
			5678: NewNamespace(species.CLONE_NEWNET, 5678, ""),
			1234: NewNamespace(species.CLONE_NEWNET, 1234, ""),
		}
		dr := DiscoveryResult{}
		dr.Namespaces[NetNS] = nsmap
		sortedns := dr.SortedNamespaces(NetNS)
		Expect(sortedns).To(HaveLen(2))
		Expect(sortedns[0].ID()).To(Equal(species.NamespaceID(1234)))
		Expect(sortedns[1].ID()).To(Equal(species.NamespaceID(5678)))
	})

	It("sorts namespace lists", func() {
		nslist := []Namespace{
			NewNamespace(species.CLONE_NEWUSER, 5678, ""),
			NewNamespace(species.CLONE_NEWUSER, 1234, ""),
		}
		sortedns := SortNamespaces(nslist)
		Expect(sortedns).To(HaveLen(2))
		Expect(sortedns[0].ID()).To(Equal(species.NamespaceID(1234)))
		Expect(sortedns[1].ID()).To(Equal(species.NamespaceID(5678)))

		sortedhns := SortChildNamespaces([]Hierarchy{nslist[0].(Hierarchy), nslist[1].(Hierarchy)})
		Expect(sortedhns).To(HaveLen(2))
		Expect(sortedhns[0].(Namespace).ID()).To(Equal(species.NamespaceID(1234)))
		Expect(sortedhns[1].(Namespace).ID()).To(Equal(species.NamespaceID(5678)))
	})

	It("rejects finding roots for plain namespaces", func() {
		// We only need to run a simplified discovery on processes, but
		// nothing else.
		opts := NoDiscovery
		opts.SkipProcs = false
		opts.NamespaceTypes = species.CLONE_NEWNET
		allns := Discover(opts)
		Expect(func() { rootNamespaces(allns.Namespaces[NetNS]) }).To(Panic())
	})

	It("returns namespaces in correct slots, implementing correct interfaces", func() {
		allns := Discover(FullDiscovery)
		for _, nstype := range TypeIndexLexicalOrder {
			for _, ns := range allns.Namespaces[nstype] {
				Expect(TypesByIndex[nstype]).To(Equal(ns.Type()))
				Expect(ns.(Namespace)).NotTo(BeNil())
				switch nstype {
				case PIDNS:
					Expect(ns.(Hierarchy)).NotTo(BeNil())
				case UserNS:
					Expect(ns.(Hierarchy)).NotTo(BeNil())
					Expect(ns.(Ownership)).NotTo(BeNil())
				}
			}
		}
	})

})
