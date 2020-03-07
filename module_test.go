package lxkns

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rxtst "github.com/thediveo/gons/reexec/testing"
)

func TestMain(m *testing.M) {
	// Ensure that the registered handler is run in the re-executed child.
	// This won't trigger the handler while we're in the parent. We're using
	// gons' very special coverage profiling support for re-execution.
	mm := &rxtst.M{M: m}
	os.Exit(mm.Run())
}

func TestLinuxKernelNamespaces(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns package")
}
