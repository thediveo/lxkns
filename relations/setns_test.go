package relations

import (
	"fmt"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstypes"
	"github.com/thediveo/testbasher"
)

var _ = Describe("Set Namespaces", func() {

	It("goes into other namespaces", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("newnetns", `
unshare -Unr $netns
`)
		scripts.Script("netns", `
echo "\"/proc/$$/ns/user\""
echo "\"/proc/$$/ns/net\""
process_namespaceid net
read # wait for test to proceed()
`)
		cmd := scripts.Start("newnetns")
		defer cmd.Close()

		var usernsref, netnsref NamespacePath
		var netnsid nstypes.NamespaceID
		cmd.Decode(&usernsref)
		cmd.Decode(&netnsref)
		cmd.Decode(&netnsid)

		result := make(chan nstypes.NamespaceID)
		Expect(Go(func() {
			id, _ := NamespacePath(
				fmt.Sprintf("/proc/%d/ns/net", syscall.Gettid())).
				ID()
			result <- id
		}, /*usernsref, */ netnsref)).NotTo(HaveOccurred())
		Expect(<-result).To(Equal(netnsid))
	})

})
