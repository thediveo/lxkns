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
	"github.com/thediveo/lxkns/species"
)

var proc1 = &lxkns.Process{
	PID:        1,
	PPID:       0,
	Cmdline:    []string{"/sbin/domination", "--world"},
	Name:       "(init)",
	Starttime:  123,
	Namespaces: namespaces,
}

const proc1JSON = `{
	"namespaces": {
	  "mnt": 66610,
	  "cgroup": 66611,
	  "uts": 66612,
	  "ipc": 66613,
	  "user": 66614,
	  "pid": 66615,
	  "net": 66616
	},
	"pid": 1,
	"ppid": 0,
	"name": "(init)",
	"cmdline": [
	  "/sbin/domination",
	  "--world"
	],
	"starttime": 123
}`

var proc2 = &lxkns.Process{
	PID:        666,
	PPID:       proc1.PID,
	Cmdline:    []string{"/sbin/fool"},
	Name:       "fool",
	Starttime:  666666,
	Namespaces: namespaces,
}

const proc2JSON = `{
	"namespaces": {
	  "mnt": 66610,
	  "cgroup": 66611,
	  "uts": 66612,
	  "ipc": 66613,
	  "user": 66614,
	  "pid": 66615,
	  "net": 66616
	},
	"pid": 666,
	"ppid": 1,
	"name": "fool",
	"cmdline": [
	  "/sbin/fool"
	],
	"starttime": 666666
}`

var namespaces = lxkns.NamespacesSet{
	lxkns.NewNamespace(species.CLONE_NEWNS, species.NamespaceID{Dev: 666, Ino: 66610}, ""),
	lxkns.NewNamespace(species.CLONE_NEWCGROUP, species.NamespaceID{Dev: 666, Ino: 66611}, ""),
	lxkns.NewNamespace(species.CLONE_NEWUTS, species.NamespaceID{Dev: 666, Ino: 66612}, ""),
	lxkns.NewNamespace(species.CLONE_NEWIPC, species.NamespaceID{Dev: 666, Ino: 66613}, ""),
	lxkns.NewNamespace(species.CLONE_NEWUSER, species.NamespaceID{Dev: 666, Ino: 66614}, ""),
	lxkns.NewNamespace(species.CLONE_NEWPID, species.NamespaceID{Dev: 666, Ino: 66615}, ""),
	lxkns.NewNamespace(species.CLONE_NEWNET, species.NamespaceID{Dev: 666, Ino: 66616}, ""),
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

var _ = Describe("JSON", func() {

	It("marshals NamespacesSetReferences", func() {
		j, err := json.Marshal((*NamespacesSetReferences)(&namespaces))
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(namespacesJSON))
	})

	It("unmarshals NamespacesSetReferences", func() {
		allns := lxkns.NewAllNamespaces()
		nsrefs := &NamespacesSetReferences{}

		// This must NOT work...
		Expect(func() { _ = json.Unmarshal([]byte(`{"mnt":12345}`), nsrefs) }).To(Panic())

		// Expect correct failure ;)
		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": "foobar"}`), allns)).To(HaveOccurred())
		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": 123}`), allns)).To(HaveOccurred())

		// Do we correctly all the references and are they also entered into
		// the namespace dictionary for later reuse?
		Expect(nsrefs.unmarshalJSON([]byte(namespacesJSON), allns)).NotTo(HaveOccurred())
		for _, i := range []struct {
			idx lxkns.NamespaceTypeIndex
			len int
		}{
			{lxkns.MountNS, 1},
			{lxkns.CgroupNS, 1},
			{lxkns.UTSNS, 1},
			{lxkns.IPCNS, 1},
			{lxkns.UserNS, 1},
			{lxkns.PIDNS, 1},
			{lxkns.NetNS, 1},
			{lxkns.TimeNS, 0},
		} {
			Expect(allns[i.idx]).To(HaveLen(i.len),
				"wrong length of namespace map type %s", lxkns.TypesByIndex[i.idx])
		}

		// Does a second unmarshalling reuse the already known namespace
		// objects?
		nsrefs2 := &NamespacesSetReferences{}
		Expect(nsrefs2.unmarshalJSON([]byte(`{"mnt": 66610}`), allns)).NotTo(HaveOccurred())
		Expect(nsrefs2[lxkns.MountNS]).To(BeIdenticalTo(nsrefs[lxkns.MountNS]))
	})

	It("un/marshals Process", func() {
		dummy := &Process{}

		// This must NOT work...
		Expect(func() { _ = json.Unmarshal([]byte(`{"foobar":"foobar"}`), dummy) }).To(Panic())

		// Check correct failure...
		Expect(dummy.unmarshalJSON([]byte(`"foobar"`), nil)).To(HaveOccurred())
		Expect(dummy.unmarshalJSON([]byte(`{}`), nil)).To(HaveOccurred())
		Expect(dummy.unmarshalJSON([]byte(`{"pid":1,"namespaces":{"foobar":666}}`), nil)).To(HaveOccurred())

		// First establish that serialization works as expected...
		j, err := json.Marshal((*Process)(proc1))
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(proc1JSON))
		// Next, deserialize the correct JSON textural serialization again...
		allns := lxkns.NewAllNamespaces()
		p := &Process{}
		Expect(p.unmarshalJSON(j, allns)).NotTo(HaveOccurred())
		// ...but how to we know it deserialization worked as expected? By
		// serializing the deserialized Process object again, seeing if we end
		// up with the same JSON textual representation.
		j2, err := json.Marshal(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(j2).To(MatchJSON(j))
	})

	It("marshals ProcessTable", func() {
		pt := &ProcessTable{
			ProcessTable: lxkns.ProcessTable{proc1.PID: proc1, proc2.PID: proc2},
		}
		j, err := json.Marshal(pt)
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(`{"1":` + proc1JSON + `,"666":` + proc2JSON + `}`))
	})

	It("unmarshals ProcessTable", func() {
		// Check correct failure...
		dummy := &ProcessTable{}
		Expect(json.Unmarshal([]byte(`"foobar"`), dummy)).To(HaveOccurred())
		Expect(json.Unmarshal([]byte(`{"1":{"namespaces":{"foobar":666}}}`), dummy)).To(HaveOccurred())

		// To unmarshal ... we need to ... marshal first!
		pt := &ProcessTable{
			ProcessTable: lxkns.ProcessTable{proc1.PID: proc1, proc2.PID: proc2},
		}
		j, err := json.Marshal(pt)
		Expect(err).NotTo(HaveOccurred())

		// Set up an empty process table with a suitable namespace dictionary,
		// and then try to unmarshal the JSON we've just marshalled before.
		pt2 := &ProcessTable{Namespaces: lxkns.NewAllNamespaces()}
		Expect(json.Unmarshal(j, pt2)).NotTo(HaveOccurred())
		Expect(pt2.ProcessTable).To(HaveLen(len(pt.ProcessTable)))
		// Ensure that the namespace dictionary has been correctly updated and
		// that processes with the same namespaces share the same namespace
		// objects.
		Expect(pt2.Namespaces[lxkns.MountNS]).To(HaveLen(1))
		Expect(pt2.Namespaces[lxkns.TimeNS]).To(HaveLen(0))
		Expect(pt2.ProcessTable[proc1.PID].Namespaces[lxkns.CgroupNS]).NotTo(BeNil())
		Expect(pt2.ProcessTable[proc1.PID].Namespaces[lxkns.CgroupNS]).To(
			BeIdenticalTo(pt2.ProcessTable[proc2.PID].Namespaces[lxkns.CgroupNS]))
	})

})
