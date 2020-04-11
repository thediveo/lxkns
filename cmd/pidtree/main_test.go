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

package main

import (
	"bytes"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstypes"
	t "github.com/thediveo/lxkns/nstypes"
	"github.com/thediveo/testbasher"
)

var _ = Describe("renders PID trees and branches", func() {

	var scripts testbasher.Basher
	var cmd *testbasher.TestCommand
	var pidnsid t.NamespaceID
	var initpid, leafpid lxkns.PIDType

	BeforeEach(func() {
		cmd = nil
		scripts = testbasher.Basher{}
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Upmfr $stage2
`)
		scripts.Script("stage2", `
mount -t proc proc /proc
process_namespaceid pid # print ID of new PID namespace.
echo "$$"
(echo $BASHPID && read)
`)
		cmd = scripts.Start("main")
		cmd.Decode(&pidnsid)
		cmd.Decode(&initpid)
		Expect(initpid).To(Equal(lxkns.PIDType(1)))
		cmd.Decode(&leafpid)
	})

	AfterEach(func() {
		if cmd != nil {
			cmd.Close()
		}
		scripts.Done()
	})

	It("renders a PID tree", func() {
		out := bytes.Buffer{}
		renderPIDTreeWithNamespaces(&out)
		tree := out.String()
		Expect(tree).To(MatchRegexp(fmt.Sprintf(`
(?m)^[│ ]+└─ "unshare" \(\d+\)
[│ ]+└─ pid:\[%d\], owned by UID %d \(".*"\)
[│ ]+└─ "stage2.sh" \(\d+/1\)
[│ ]+└─ "stage2.sh" \(\d+/%d\)$`,
			pidnsid, os.Geteuid(), leafpid)))
	})

	It("renders only a branch", func() {
		out := bytes.Buffer{}
		renderPIDBranch(&out, lxkns.PIDType(initpid), nstypes.NamespaceID(pidnsid))
		tree := out.String()
		Expect(tree).To(MatchRegexp(fmt.Sprintf(`
(?m)^ +└─ pid:\[%d\], owned by UID %d \(".*"\)
\ +└─ "stage2.sh" \(\d+/1\)
$`,
			pidnsid, os.Geteuid())))
	})

})
