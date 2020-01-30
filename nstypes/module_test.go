package nstypes

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNamespaceTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns/nstypes package")
}
