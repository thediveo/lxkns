// Copyright 2023 Harald Albrecht.
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
	"strconv"
	"strings"
	"time"

	"github.com/thediveo/spacetest/mntns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("task-based pauser", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("mounts a mount namespace", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		_, procfsroot := mntns.NewTransient()
		pid := Successful(strconv.ParseUint(strings.Split(procfsroot, "/")[2], 10, 32))

		p := Successful(newPauseTask(fmt.Sprintf("/proc/%d/ns/mnt", pid)))
		p.Close()
	})

})
