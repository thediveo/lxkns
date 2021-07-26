// Copyright 2021 Harald Albrecht.
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

package mounteneer

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/errxpect"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/testbasher"
)

var _ = Describe("mounteneer", func() {

	It("returns error for non-existing sandbox binary", func() {
		Errxpect(newPauseProcess("/not-existing", "/proc/self/ns/mnt")).To(MatchError(
			MatchRegexp(`cannot start pause process: .* /not-existing: no such file or directory`)))
	})

	It("returns sandbox errors", func() {
		Errxpect(newPauseProcess("/proc/self/exe", "/proc/self/non-existing")).To(MatchError(
			MatchRegexp(`invalid mount namespace reference .* No such file or directory`)))
	})

	It("mounts a mount namespace", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -mr $stage2
`)
		scripts.Script("stage2", `
echo $$
read
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var pid int
		cmd.Decode(&pid)

		p, err := NewPauseProcess(fmt.Sprintf("/proc/%d/ns/mnt", pid))
		Expect(err).NotTo(HaveOccurred())
		_ = p.Process.Kill()
	})

})
