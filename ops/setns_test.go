// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ops

import (
	"fmt"
	"os"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstypes"
	"github.com/thediveo/testbasher"
)

var _ = Describe("Set Namespaces", func() {

	It("goes into other namespaces", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("newnetns", `
echo "\"/proc/$$/ns/net\""
process_namespaceid net
read # wait for test to proceed()
`)
		cmd := scripts.Start("newnetns")
		defer cmd.Close()

		var netnsref NamespacePath
		var netnsid nstypes.NamespaceID
		cmd.Decode(&netnsref)
		cmd.Decode(&netnsid)

		result := make(chan nstypes.NamespaceID)
		Expect(Go(func() {
			id, _ := NamespacePath(
				fmt.Sprintf("/proc/%d/ns/net", syscall.Gettid())).
				ID()
			result <- id
		}, netnsref)).NotTo(HaveOccurred())
		Expect(<-result).To(Equal(netnsid))
	})

})
