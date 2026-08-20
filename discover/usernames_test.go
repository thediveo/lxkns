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
	"bytes"
	"log/slog"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("maps UIDs", func() {

	BeforeEach(func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("returns same information as library queries", func() {
		myuid := os.Getuid()
		u, err := user.LookupId(strconv.FormatUint(uint64(myuid), 10))
		Expect(err).NotTo(HaveOccurred())
		Expect(u).NotTo(BeNil())
		myusername := u.Username

		u, err = user.LookupId("0")
		Expect(err).NotTo(HaveOccurred())
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
		mnteer := Successful(mountineer.New([]string{"/proc/1/ns/mnt"}, nil))
		defer mnteer.Close()
		useruidmap := map[string]uint32{}
		for line := range bytes.Lines(Successful(mnteer.ReadFile("/etc/passwd"))) {
			fields := strings.Split(string(line), ":")
			if len(fields) < 3 {
				continue
			}
			useruidmap[fields[0]] = uint32(Successful(strconv.Atoi(fields[2])))
		}

		hostuidusermap := UidUsernameMap{}
		for user, uid := range useruidmap {
			hostuidusermap[uid] = user
		}

		usernames := DiscoverUserNames(allns.Namespaces)
		Expect(usernames).To(HaveLen(len(useruidmap)))
		for uid, username := range hostuidusermap {
			Expect(usernames[uid]).To(Equal(username), "missing uid %d: %q", uid, username)
		}
	})

})
