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

var proc2 = &lxkns.Process{
	PID:        666,
	PPID:       proc1.PID,
	Cmdline:    []string{"/sbin/fool"},
	Name:       "fool",
	Starttime:  666666,
	Namespaces: namespaces,
}

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

var _ = Describe("JSON", func() {

	It("unmarshals TypedNamespacesSet", func() {
		// With context...
		allns := lxkns.NewAllNamespaces()
		nsrefs := &TypedNamespacesSet{}

		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": "foobar"}`), allns)).To(HaveOccurred())
		Expect(nsrefs.unmarshalJSON([]byte(`{"foobar": 123}`), allns)).To(HaveOccurred())

		Expect(nsrefs.unmarshalJSON([]byte(`{
			"mnt": 66610,
			"cgroup": 66611,
			"uts": 66612,
			"ipc": 66613,
			"user": 66614,
			"pid": 66615,
			"net": 66616
		  }`), allns)).NotTo(HaveOccurred())
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
	})

	It("un/marshals Process", func() {
		dummy := &Process{}
		Expect(json.Unmarshal([]byte(`"foobar"`), dummy)).To(HaveOccurred())
		Expect(json.Unmarshal([]byte(`{}`), dummy)).To(HaveOccurred())
		Expect(json.Unmarshal([]byte(`{"pid":1,"namespaces":{"foobar":666}}`), dummy)).To(HaveOccurred())

		// First establish that serialization works as expected...
		j, err := json.Marshal(&Process{Process: proc1})
		Expect(err).NotTo(HaveOccurred())
		Expect(j).To(MatchJSON(`{
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
		  }`))
		// Next, deserialize the correct JSON textural serialization again...
		p := &Process{AllNamespaces: lxkns.NewAllNamespaces()}
		Expect(json.Unmarshal(j, p)).NotTo(HaveOccurred())
		// ...but how to we know it deserialization worked as expected? By
		// serializing the deserialized Process object again, seeing if we end
		// up with the same JSON textual representation.
		j2, err := json.Marshal(p)
		Expect(err).NotTo(HaveOccurred())
		Expect(j2).To(MatchJSON(j))
	})

	/*
		It("ProcessTable un/marshalling", func() {
			proc1 := &lxkns.Process{
				PID:       1,
				PPID:      0,
				Cmdline:   []string{"/sbin/domination", "--world"},
				Name:      "(init)",
				Starttime: 123,
				// TODO: Namespaces
			}
			proc2 := &lxkns.Process{
				PID:       666,
				PPID:      proc1.PID,
				Cmdline:   []string{"/sbin/fool"},
				Name:      "fool",
				Starttime: 666666,
				// TODO: Namespaces
			}
			procs := ProcessTable{
				proc1.PID: proc1,
				proc2.PID: proc2,
			}
			_, err := json.Marshal(procs)
			Expect(err).NotTo(HaveOccurred())
			// FIXME: Expect(j).To(MatchJSON(`""`))
		})
	*/

})
