// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package portable

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("portable reference integration", func() {

	It("opens portable (network) namespace reference and runs a sub-process in it", func() {
		// We need to create a new network namespace which we later want to
		// enter in a separate Go routine. Unfortunately, while we could create
		// a new network namespace when creating a new user namespace first, we
		// won't be allowed to enter another user namespace because we're
		// already OS multi-threaded.
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -n $stage2
`)
		scripts.Script("stage2", `
process_namespaceid net
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()

		netnsid := nstest.CmdDecodeNSId(cmd)

		netns, closer, err := PortableReference{ID: netnsid, Type: species.CLONE_NEWNET}.Open()
		Expect(err).To(Succeed())
		defer closer()
		res, err := ops.Execute(func() interface{} {
			cmd := exec.Command("ls", "-l", "/proc/self/ns/net")
			out, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}
			return out
		}, netns)
		Expect(err).To(Succeed())
		Expect(res).To(BeAssignableToTypeOf([]byte{}))
		b, _ := res.([]byte)
		Expect(string(b)).To(MatchRegexp(fmt.Sprintf(`net:\[%d\]`, netnsid.Ino)))
	})

})
