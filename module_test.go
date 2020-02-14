package lxkns

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/gons/reexec"
)

func TestLinuxKernelNamespaces(t *testing.T) {
	reexec.CheckAction()
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns package")
}
