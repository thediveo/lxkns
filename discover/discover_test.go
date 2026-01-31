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

package discover

import (
	"log/slog"
	"time"

	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("Discover", func() {

	BeforeEach(func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("discovers nothing unless told so", func() {
		allns := Namespaces()
		for _, nsmap := range allns.Namespaces {
			Expect(nsmap).To(BeEmpty())
		}
	})

	It("ignores a nil option", func() {
		Expect(func() { _ = Namespaces(nil) }).NotTo(Panic())
	})

	It("sorts namespace maps", func() {
		nsmap := model.NamespaceMap{
			species.NamespaceID{Dev: 1, Ino: 5678}: namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NamespaceID{Dev: 1, Ino: 5678}, ""),
			species.NamespaceID{Dev: 1, Ino: 1234}: namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NamespaceID{Dev: 1, Ino: 1234}, ""),
		}
		dr := Result{}
		dr.Namespaces[model.NetNS] = nsmap
		sortedns := dr.SortedNamespaces(model.NetNS)
		Expect(sortedns).To(HaveLen(2))
		Expect(sortedns[0].ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 1234}))
		Expect(sortedns[1].ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 5678}))
	})

	It("sorts namespace lists", func() {
		nslist := []model.Namespace{
			namespaces.NewWithSimpleRef(species.CLONE_NEWUSER, species.NamespaceID{Dev: 1, Ino: 5678}, ""),
			namespaces.NewWithSimpleRef(species.CLONE_NEWUSER, species.NamespaceID{Dev: 1, Ino: 1234}, ""),
		}
		sortedns := SortNamespaces(nslist)
		Expect(sortedns).To(HaveLen(2))
		Expect(sortedns[0].ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 1234}))
		Expect(sortedns[1].ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 5678}))

		sortedhns := SortChildNamespaces([]model.Hierarchy{nslist[0].(model.Hierarchy), nslist[1].(model.Hierarchy)})
		Expect(sortedhns).To(HaveLen(2))
		Expect(sortedhns[0].(model.Namespace).ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 1234}))
		Expect(sortedhns[1].(model.Namespace).ID()).To(Equal(species.NamespaceID{Dev: 1, Ino: 5678}))
	})

	It("rejects finding roots for plain namespaces", func() {
		// We only need to run a simplified discovery on processes, but
		// nothing else.
		allns := Namespaces(FromProcs(), WithNamespaceTypes(species.CLONE_NEWNET))
		Expect(func() { rootNamespaces(allns.Namespaces[model.NetNS]) }).To(Panic())
	})

	It("returns namespaces in correct slots, implementing correct interfaces", func() {
		allns := Namespaces(WithStandardDiscovery())
		for _, nstype := range model.TypeIndexLexicalOrder {
			for _, ns := range allns.Namespaces[nstype] {
				Expect(model.TypesByIndex[nstype]).To(Equal(ns.Type()))
				Expect(ns).NotTo(BeNil())
				switch nstype {
				case model.PIDNS:
					Expect(ns.(model.Hierarchy)).NotTo(BeNil())
				case model.UserNS:
					Expect(ns.(model.Hierarchy)).NotTo(BeNil())
					Expect(ns.(model.Ownership)).NotTo(BeNil())
				}
			}
		}
	})

})
