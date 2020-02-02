package relations

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRelations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lxkns/relations package")
}
