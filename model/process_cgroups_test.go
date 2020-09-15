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

package model

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("cgrouping", func() {

	It("finds hidden hierarchical user namespaces", func() {
		_, err := cgroupMountpath("cgroupv666fooobarcontroller")
		Expect(err).To(MatchError(MatchRegexp(`controller .+ not mounted`)))

		cpu, err := cgroupMountpath("cpu")
		Expect(err).To(Succeed())
		Expect(cpu).To(HavePrefix("/sys/fs/cgroup/"))

		cpuacct, err := cgroupMountpath("cpuacct")
		Expect(err).To(Succeed())
		Expect(cpuacct).To(HavePrefix("/sys/fs/cgroup/"))

		Expect(cpu).To(Equal(cpuacct))
	})

	It("finds control groups of processes", func() {
		procs := NewProcessTable()
		Expect(procs).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
			"Controlgroup": Not(BeEmpty()),
		}))))
	})

})
