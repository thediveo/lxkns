package lxkns

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLinuxKernelNamespaces(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns suite")
}
