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

package caps

import (
	"os"

	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("effective caps", func() {

	It("reads nil capabilities set from bad status file", func() {
		f, err := os.Open("./test/status-bad")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = f.Close() }()
		Expect(statusEffectiveCaps(f)).To(BeNil())
	})

	It("reads nil capabilities set from corrupt status file", func() {
		f, err := os.Open("./test/status-corrupt")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = f.Close() }()
		Expect(statusEffectiveCaps(f)).To(BeNil())
	})

	It("reads capabilities set from good status file", func() {
		f, err := os.Open("./test/status-good")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = f.Close() }()
		caps := statusEffectiveCaps(f)
		Expect(caps).To(HaveLen(2))
		Expect(caps[0]).To(Equal(uint32(0xffffffff)))
		Expect(caps[1]).To(Equal(uint32(0x3f)))
	})

	It("reads nil capabilities set for non-existing process", func() {
		Expect(processEffectiveCaps(model.PIDType(-1))).To(BeNil())
	})

	It("reads effective capabilities set from process status", func() {
		Expect(processEffectiveCaps(model.PIDType(os.Getpid()))).ToNot(BeNil())
	})

	It("returns capabilities of init process", func() {
		caps := ProcessCapabilities(model.PIDType(1))
		Expect(len(caps)).To(BeNumerically(">=", 2))
		Expect(caps).To(ContainElement("cap_sys_ptrace"))
	})

})
