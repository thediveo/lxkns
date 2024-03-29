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

package mountineer

import (
	"fmt"
	"os"
	"time"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("process-based pauser", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("returns error for non-existing sandbox binary", func() {
		Expect(newPauseProcessWithBinary("/not-existing", "/proc/self/ns/mnt", "")).
			Error().To(MatchError(MatchRegexp(
			`cannot start pause process, reason: .* /not-existing: no such file or directory`)))
	})

	It("returns sandbox errors", func() {
		Expect(newPauseProcessWithBinary("/proc/self/exe", "/proc/self/non-existing", "")).
			Error().To(MatchError(MatchRegexp(
			`invalid mount namespace reference .* No such file or directory`)))
	})

	It("mounts a mount namespace", func() {
		if os.Getuid() != 0 {
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

		p, err := newPauseProcess(
			fmt.Sprintf("/proc/%d/ns/mnt", pid),
			fmt.Sprintf("/proc/%d/ns/user", pid))
		Expect(err).NotTo(HaveOccurred())
		Expect(p.PID()).NotTo(BeZero())
		p.Close()
	})

})
