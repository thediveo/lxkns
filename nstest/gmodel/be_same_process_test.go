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

package gmodel

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var proc1 = &model.Process{
	PID:        1,
	PPID:       0,
	Cmdline:    []string{"/sbin/domination", "--world"},
	Name:       "(init)",
	Starttime:  123,
	Namespaces: namespaceset,
}

func init() {
	proc1.Children = append(proc1.Children, proc2)
	proc2.Parent = proc1
}

var proc2 = &model.Process{
	PID:        666,
	PPID:       proc1.PID,
	Cmdline:    []string{"/sbin/fool"},
	Name:       "fool",
	Starttime:  666666,
	Namespaces: namespaceset,
}

var namespaceset = model.NamespacesSet{
	namespaces.NewWithSimpleRef(species.CLONE_NEWNS, species.NamespaceID{Dev: 666, Ino: 66610}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWCGROUP, species.NamespaceID{Dev: 666, Ino: 66611}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWUTS, species.NamespaceID{Dev: 666, Ino: 66612}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWIPC, species.NamespaceID{Dev: 666, Ino: 66613}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWUSER, species.NamespaceID{Dev: 666, Ino: 66614}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWPID, species.NamespaceID{Dev: 666, Ino: 66615}, ""),
	namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NamespaceID{Dev: 666, Ino: 66616}, ""),
	nil,
}

var _ = Describe("Process", func() {

	It("handles mistakes", func() {
		Expect(BeSameTreeProcess(nil).Match(nil)).Error().To(MatchError(MatchRegexp(`use BeNil()`)))
		Expect(BeSameTreeProcess(proc1).Match("foo")).Error().To(
			MatchError(MatchRegexp(`expects a model.Process, not a string`)))
		Expect(BeSameTreeProcess("foo").Match(proc1)).Error().To(
			MatchError(MatchRegexp(`must be passed a model.Process, not a string`)))
	})

	It("matches", func() {
		Expect(proc1).NotTo(BeSameTreeProcess(nil))

		// Expect BeSameTreeProcess to "unpack" a pointer'ed Process.
		Expect(*proc1).To(BeSameTreeProcess(proc1))

		// Expect that the parents are checked to be similar. This is an ugly
		// hack of a test: we modify the parent reference, but not the PPID ;)
		proc := *proc1
		proc.Parent = proc1
		Expect(proc).NotTo(BeSameTreeProcess(*proc1))
		Expect(proc).To(BeSameProcess(*proc1))

		// Expect that the children are also checked to be similar.
		proc = *proc1
		proc.Children = append(proc.Children, proc1)
		Expect(proc).NotTo(BeSameTreeProcess(*proc1))

		// Expect that the same (similar) namespaces have been joined.
		proc = *proc1
		proc.Namespaces[model.UserNS] = nil
		Expect(proc).NotTo(BeSameTreeProcess(*proc1))
	})

	It("matches namespace references", func() {
		Expect(sameRef(nil, nil)).To(BeTrue())
		Expect(sameRef(model.NamespaceRef{}, model.NamespaceRef{})).To(BeTrue())
		Expect(sameRef(model.NamespaceRef{"/foo"}, model.NamespaceRef{"/foo"})).To(BeTrue())
		Expect(sameRef(model.NamespaceRef{"/foo"}, model.NamespaceRef{"/bar"})).NotTo(BeTrue())
		Expect(sameRef(model.NamespaceRef{"/foo", "/bar"}, model.NamespaceRef{"/foo"})).NotTo(BeTrue())
	})

})
