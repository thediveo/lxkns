package lxkns

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLinuxKernelNamespaces(t *testing.T) {
	ExecReexecAction()
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns package")
}
