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

package lxkns

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("maps UIDs", func() {

	It("returns same information as direct query", func() {
		myuid := os.Getuid()
		u, err := user.LookupId(strconv.FormatUint(uint64(myuid), 10))
		Expect(err).To(Succeed())
		Expect(u).NotTo(BeNil())
		myusername := u.Username

		u, err = user.LookupId("0")
		Expect(err).To(Succeed())
		Expect(u).NotTo(BeNil())
		rootname := u.Username

		unames := userNamesOracle([]uint32{uint32(myuid), 0})
		Expect(unames).To(HaveKeyWithValue(uint32(0), rootname))
		Expect(unames).To(HaveKeyWithValue(uint32(myuid), myusername))

		unames = map[uint32]string{}
		err = ReexecIntoActionEnv(
			"discover-uid-names",
			[]model.Namespace{},
			[]string{fmt.Sprintf("UIDS=[0,%d,65535]", myuid)},
			&unames)
		Expect(err).To(Succeed())
		Expect(unames).To(HaveKeyWithValue(uint32(0), rootname))
		Expect(unames).To(HaveKeyWithValue(uint32(myuid), myusername))
	})

	It("handles failing action", func() {
		Expect(readUidNames).To(Panic())

		usernames := map[uint32]string{}
		err := ReexecIntoActionEnv(
			"discover-uid-names",
			[]model.Namespace{},
			[]string{"UIDS=42"},
			&usernames)
		Expect(err).To(MatchError(MatchRegexp(`cannot unmarshal number into Go value`)))
	})

	FIt("switches into initial namespace", func() {
		allns := Discover(FullDiscovery)
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

		usernames := queryUserNames([]uint32{0, 1000}, allns.Namespaces)
		Expect(len(usernames)).To(BeNumerically(">=", 1))
	})

})
