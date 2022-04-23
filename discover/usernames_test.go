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

package discover

import (
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("maps UIDs", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("returns same information as library queries", func() {
		myuid := os.Getuid()
		u, err := user.LookupId(strconv.FormatUint(uint64(myuid), 10))
		Expect(err).To(Succeed())
		Expect(u).NotTo(BeNil())
		myusername := u.Username

		u, err = user.LookupId("0")
		Expect(err).To(Succeed())
		Expect(u).NotTo(BeNil())
		rootname := u.Username

		unames := userNamesFromPasswd(etcpasswd)
		Expect(unames).To(HaveKeyWithValue(uint32(0), rootname))
		Expect(unames).To(HaveKeyWithValue(uint32(myuid), myusername))
	})

	It("switches into initial namespace and reads user names", func() {
		// This test is unusual, as we can carry it out only when we're inside
		// a separate mount namespace, so we can't immediately see the users
		// on the host system itself. We need some checks to ensure that we're
		// going to test things in the correct setup.
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		allns := Namespaces(WithStandardDiscovery())
		if _, ok := allns.Processes[1]; !ok {
			Skip("needs root capabilities and PID=host")
		}
		mymntns := allns.Processes[1].Namespaces[model.MountNS]
		initialmntns := allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.MountNS]
		if mymntns == initialmntns {
			Skip("needs container with different mount namespace")
		}
		if initialmntns == nil {
			Skip("needs PID=host")
		}

		// In order to check the data we want to discover, we need an
		// independent second view. Now, that's a job for safety, not for
		// reliability.
		scripts := testbasher.Basher{}
		scripts.Common(nstest.NamespaceUtilsScript)
		// Remember: we're here now in a container with root privileges. And
		// this needs awk in the host. And then there are probably differences
		// between nsenter made by Alpine(hmpf) and nsenter on the host system
		// in terms of their CLI flags, so we need to detect the CLI flag
		// variant to use...
		scripts.Script("main", `
ENTERMNT=$(nsenter -h 2>&1 | grep -q -e "--mnt" && echo "--mnt" || echo "-m")
nsenter -t 1 ${ENTERMNT} -- /usr/bin/awk -F: 'BEGIN{printf "{"}{printf "\"%s\":%s,",$1,$3}END{printf "\"guardian-fooobar\":666}\n"}' /etc/passwd
read
`)
		scriptscmd := scripts.Start("main")
		var useruidmap map[string]uint32
		scriptscmd.Decode(&useruidmap)
		Expect(useruidmap).To(HaveKeyWithValue("guardian-fooobar", uint32(666)))
		hostuidusermap := UidUsernameMap{}
		for user, uid := range useruidmap {
			if uid != 666 {
				hostuidusermap[uint32(uid)] = user
			}
		}
		scriptscmd.Close()
		scripts.Done()

		usernames := DiscoverUserNames(allns.Namespaces)
		Expect(len(usernames)).To(Equal(len(useruidmap) - 1))
		for uid, username := range hostuidusermap {
			Expect(usernames[uid]).To(Equal(username), "missing uid %d: %q", uid, username)
		}
	})

})
