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
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("effective caps", func() {

	It("reads empty octet string from bad status file", func() {
		f, err := os.Open("./test/status-bad")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		b := statusEffectiveCaps(f)
		Expect(b).To(HaveLen(0))
	})

	It("reads empty octet string from corrupt status file", func() {
		f, err := os.Open("./test/status-corrupt")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		b := statusEffectiveCaps(f)
		Expect(b).To(HaveLen(0))
	})

	It("reads octet string from good status file", func() {
		f, err := os.Open("./test/status-good")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		b := statusEffectiveCaps(f)
		Expect(b).To(HaveLen(8))
		Expect(b[0]).To(Equal(byte(0xff)))
		Expect(b[4]).To(Equal(byte(0x3f)))
	})

	It("reads octet string from process status", func() {
		b := processEffectiveCaps(model.PIDType(-1))
		Expect(b).To(BeEmpty())
		b = processEffectiveCaps(model.PIDType(os.Getpid()))
		Expect(b).ToNot(BeEmpty())
	})

	It("returns caps of init process", func() {
		caps := ProcessCapabilities(model.PIDType(1))
		Expect(len(caps)).To(BeNumerically(">=", 8))
		Expect(caps).To(ContainElement("cap_sys_ptrace"))
	})

	It("converts cap byte strings to cap names", func() {
		caps := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}
		names := capsToNames(caps)
		Expect(names).To(HaveLen(2))
		Expect(names).To(ContainElement(CapNames[0]))
		Expect(names).To(ContainElement(fmt.Sprintf("cap_%d", len(caps)*8-1)))
	})

})
