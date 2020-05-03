package main

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
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
		b := processEffectiveCaps(lxkns.PIDType(-1))
		Expect(b).To(BeEmpty())
		b = processEffectiveCaps(lxkns.PIDType(os.Getpid()))
		Expect(b).ToNot(BeEmpty())
	})

	It("returns caps of init process", func() {
		caps := ProcessCapabilities(lxkns.PIDType(1))
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
