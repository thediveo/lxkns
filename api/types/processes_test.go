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
	. "github.com/onsi/gomega/gstruct"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest/gmodel"
	"github.com/thediveo/lxkns/species"
)

var proc1 = &model.Process{
	PID:          1,
	PPID:         0,
	Cmdline:      []string{"/sbin/domination", "--world"},
	Name:         "(init)",
	Starttime:    123,
	Namespaces:   namespaceset,
	Fridge:       model.ProcessThawed,
	Selffridge:   model.ProcessThawed,
	Parentfridge: model.ProcessThawed,
}

const proc1JSON = `{
	"namespaces": ` + namespacesJSON + `,
	"pid": 1,
	"ppid": 0,
	"name": "(init)",
	"cmdline": [
	  "/sbin/domination",
	  "--world"
	],
	"starttime": 123,
	"cgroup": "",
	"fridgecgroup": "",
	"fridge": "thawed",
	"selffridge": "thawed",
	"parentfridge": "thawed"
}`

var proc2 = &model.Process{
	PID:          666,
	PPID:         proc1.PID,
	Cmdline:      []string{"/sbin/fool"},
	Name:         "fool",
	Starttime:    666666,
	Namespaces:   namespaceset,
	Controlgroup: "süstem.sluice",
	FridgeCgroup: "süstem.sluice/lxkns",
	Fridge:       model.ProcessFreezing,
	Selffridge:   model.ProcessFrozen,
	Parentfridge: model.ProcessThawed,
}

const proc2JSON = `{
	"namespaces": ` + namespacesJSON + `,
	"pid": 666,
	"ppid": 1,
	"name": "fool",
	"cmdline": [
	  "/sbin/fool"
	],
	"starttime": 666666,
	"cgroup": "süstem.sluice",
	"fridgecgroup": "süstem.sluice/lxkns",
	"fridge": "freezing",
	"selffridge": "frozen",
	"parentfridge": "thawed"
}`

var namespaceset = model.NamespacesSet{
	namespaces.New(species.CLONE_NEWNS, species.NamespaceID{Dev: 666, Ino: 66610}, ""),
	namespaces.New(species.CLONE_NEWCGROUP, species.NamespaceID{Dev: 666, Ino: 66611}, ""),
	namespaces.New(species.CLONE_NEWUTS, species.NamespaceID{Dev: 666, Ino: 66612}, ""),
	namespaces.New(species.CLONE_NEWIPC, species.NamespaceID{Dev: 666, Ino: 66613}, ""),
	namespaces.New(species.CLONE_NEWUSER, species.NamespaceID{Dev: 666, Ino: 66614}, ""),
	namespaces.New(species.CLONE_NEWPID, species.NamespaceID{Dev: 666, Ino: 66615}, ""),
	namespaces.New(species.CLONE_NEWNET, species.NamespaceID{Dev: 666, Ino: 66616}, ""),
	nil,
}

const namespacesJSON = `{
	"mnt": 66610,
	"cgroup": 66611,
	"uts": 66612,
	"ipc": 66613,
	"user": 66614,
	"pid": 66615,
	"net": 66616
}`

var _ = Describe("process JSON", func() {

	It("always gets a Process from the ProcessTable", func() {
		pt := NewProcessTable(WithProcessTable(model.ProcessTable{proc1.PID: proc1}))
		p1 := pt.Get(proc1.PID)
		Expect(p1).To(BeIdenticalTo(p1))

		p2 := pt.Get(model.PIDType(666))
		Expect(p2).NotTo(BeNil())
		Expect(p2.PID).To(Equal(model.PIDType(666)))
	})

	It("marshals NamespacesSetReferences", func() {
		j, err := json.Marshal((*NamespacesSetReferences)(&namespaceset))
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(namespacesJSON))
	})

	It("unmarshals NamespacesSetReferences", func() {
		allns := NewNamespacesDict(nil)
		nsrefs := &NamespacesSetReferences{}

		// This must NOT work...
		Expect(func() { _ = json.Unmarshal([]byte(`{"mnt":12345}`), nsrefs) }).To(Panic())

		// Expect correct failure ;)
		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": "foobar"}`), allns)).To(HaveOccurred())
		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": 123}`), allns)).To(HaveOccurred())

		// Do we correctly all the references and are they also entered into
		// the namespace dictionary for later reuse?
		Expect(nsrefs.unmarshalJSON([]byte(namespacesJSON), allns)).To(Succeed())
		for _, i := range []struct {
			idx model.NamespaceTypeIndex
			len int
		}{
			{model.MountNS, 1},
			{model.CgroupNS, 1},
			{model.UTSNS, 1},
			{model.IPCNS, 1},
			{model.UserNS, 1},
			{model.PIDNS, 1},
			{model.NetNS, 1},
			{model.TimeNS, 0},
		} {
			Expect(allns.AllNamespaces[i.idx]).To(HaveLen(i.len),
				"wrong length of namespace map type %s", model.TypesByIndex[i.idx])
		}

		// Does a second unmarshalling reuse the already known namespace
		// objects?
		nsrefs2 := &NamespacesSetReferences{}
		Expect(nsrefs2.unmarshalJSON([]byte(`{"mnt": 66610}`), allns)).To(Succeed())
		Expect(nsrefs2[model.MountNS]).To(BeIdenticalTo(nsrefs[model.MountNS]))
	})

	It("un/marshals Process", func() {
		dummy := &Process{}

		// This must NOT work...
		Expect(func() { _ = json.Unmarshal([]byte(`{"foobar":"foobar"}`), dummy) }).To(Panic())

		// Check correct failure...
		Expect(dummy.unmarshalJSON([]byte(`"foobar"`), nil)).To(HaveOccurred())
		Expect(dummy.unmarshalJSON([]byte(`{}`), nil)).To(HaveOccurred())
		Expect(dummy.unmarshalJSON([]byte(`{"foobar":"foobar"}`), nil)).To(HaveOccurred())
		Expect(dummy.unmarshalJSON([]byte(`{"pid":1,"namespaces":{"foobar":666}}`), nil)).To(HaveOccurred())

		// First establish that serialization works as expected...
		j, err := json.Marshal((*Process)(proc1))
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(proc1JSON))
		// Next, deserialize the correct JSON textural serialization again...
		allns := NewNamespacesDict(nil)
		p := &Process{}
		Expect(p.unmarshalJSON(j, allns)).To(Succeed())
		// ...but how to we know it deserialization worked as expected?
		Expect((*model.Process)(p)).To(gmodel.BeSameProcess((*model.Process)(proc1)))
		// Or check by serializing the deserialized Process object again,
		// seeing if we end up with the same JSON textual representation.
		j2, err := json.Marshal(p)
		Expect(err).To(Succeed())
		Expect(j2).To(MatchJSON(j))
	})

	It("marshals ProcessTable", func() {
		pt := NewProcessTable(WithProcessTable(model.ProcessTable{proc1.PID: proc1, proc2.PID: proc2}))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())
		Expect(j).To(MatchJSON(`{"1":` + proc1JSON + `,"666":` + proc2JSON + `}`))
	})

	It("unmarshals ProcessTable", func() {
		// Check correct failure...
		dummy := NewProcessTable()
		Expect(json.Unmarshal([]byte(`"foobar"`), &dummy)).To(HaveOccurred())
		Expect(json.Unmarshal([]byte(`{"1":{"namespaces":{"foobar":666}}}`), &dummy)).To(HaveOccurred())

		// To unmarshal ... we need to ... marshal first!
		pt := NewProcessTable(WithProcessTable(model.ProcessTable{proc1.PID: proc1, proc2.PID: proc2}))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())

		// Set up an empty process table with a suitable namespace dictionary,
		// and then try to unmarshal the JSON we've just marshalled before.
		pt2 := NewProcessTable()
		Expect(json.Unmarshal(j, &pt2)).To(Succeed())
		Expect(pt2.ProcessTable).To(HaveLen(len(pt.ProcessTable)))
		// Ensure that the namespace dictionary has been correctly updated and
		// that processes with the same namespaces share the same namespace
		// objects.
		Expect(pt2.Namespaces.AllNamespaces[model.MountNS]).To(HaveLen(1))
		Expect(pt2.Namespaces.AllNamespaces[model.TimeNS]).To(HaveLen(0))
		Expect(pt2.ProcessTable[proc1.PID].Namespaces[model.CgroupNS]).NotTo(BeNil())
		Expect(pt2.ProcessTable[proc1.PID].Namespaces[model.CgroupNS]).To(
			BeIdenticalTo(pt2.ProcessTable[proc2.PID].Namespaces[model.CgroupNS]))

		// Check that preloading/priming with empty process objects works as
		// expected: the pre-existing process object must be reused, and
		// properly updated in its state.
		proc := &model.Process{}
		pt3 := NewProcessTable(WithProcessTable(model.ProcessTable{proc1.PID: proc}))
		Expect(pt3.ProcessTable[proc1.PID].PID).To(Equal(model.PIDType(0)))
		Expect(json.Unmarshal(j, &pt3)).To(Succeed())
		Expect(pt3.ProcessTable).To(HaveKey(proc1.PID))
		preproc1 := pt3.ProcessTable[proc1.PID]
		Expect(preproc1).To(BeIdenticalTo(proc))
		Expect(preproc1.PID).To(Equal(proc1.PID))
		Expect(preproc1.Children).To(HaveLen(1))
		// What did we laugh about Cobol ... and here come Gomega matchers.
		Expect(preproc1.Children).To(ContainElement(PointTo(
			MatchFields(IgnoreExtras, Fields{
				"Cmdline": Equal(proc2.Cmdline),
			})))) // ...and about bracket Lisp"ing".
	})

	It("does a full round trip without any hiccup", func() {
		pt := NewProcessTable()
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())

		jpt := NewProcessTable()
		Expect(json.Unmarshal(j, &jpt)).To(Succeed())
		Expect(jpt.ProcessTable).To(gmodel.BeSameProcessTable(pt.ProcessTable))
	})

	It("rejects bad processes", func() {
		proc := &model.Process{PID: 12345, PPID: -1, Name: "foobar"}
		pt := NewProcessTable(WithProcessTable(model.ProcessTable{proc.PID: proc}))
		j, err := json.Marshal(pt)
		Expect(err).To(Succeed())

		jpt := NewProcessTable()
		Expect(json.Unmarshal(j, &jpt)).To(HaveOccurred())
		//TODO:Expect(jpt.ProcessTable).To(gmodel.BeSameProcessTable(pt.ProcessTable))
	})

})
