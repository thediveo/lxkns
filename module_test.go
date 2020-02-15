package lxkns

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/gons/reexec"
)

func TestMain(m *testing.M) {
	// Ensure that the registered handler is run in the re-executed child. This
	// won't trigger the handler while we're in the parent, because the
	// parent's Arg[0] won't match the name of our handler.
	reexec.CheckAction()
	os.Exit(m.Run())
}

func TestLinuxKernelNamespaces(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns package")
}
