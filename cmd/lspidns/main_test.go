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
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("renders pid namespaces", func() {

	var scripts testbasher.Basher
	var cmd *testbasher.TestCommand
	var initusernsid, initpidnsid, usernsid, pidnsid species.NamespaceID

	BeforeEach(func() {
		cmd = nil
		scripts = testbasher.Basher{}
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
process_namespaceid user
process_namespaceid pid
unshare -Upmfr --mount-proc $stage2
`)
		scripts.Script("stage2", `
process_namespaceid user
process_namespaceid pid
read
`)
		cmd = scripts.Start("main")
		initusernsid = nstest.CmdDecodeNSId(cmd)
		initpidnsid = nstest.CmdDecodeNSId(cmd)
		usernsid = nstest.CmdDecodeNSId(cmd)
		pidnsid = nstest.CmdDecodeNSId(cmd)
	})

	AfterEach(func() {
		if cmd != nil {
			cmd.Close()
		}
		scripts.Done()
	})

	It("CLI --foobar fails correctly", func() {
		oldExit := osExit
		defer func() { osExit = oldExit }()
		exit := 0
		osExit = func(code int) { exit = code }
		os.Args = append(os.Args[:1], "--noengines", "--foobar")
		out := getstdout.Stdouterr(main)
		Expect(exit).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("CLI w/o args renders pid tree", func() {
		os.Args = append(os.Args[:1], "--noengines")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^pid:\[%d\] process .*$`,
			initpidnsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^[├└]─ pid:\[%d\] process .*$`,
			pidnsid.Ino)))
	})

	It("CLI w/o args renders pid tree", func() {
		os.Args = append(os.Args[:1], "--noengines", "-u")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] process .*
└─ pid:\[%d\] process .*$`,
			initusernsid.Ino, initpidnsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^   [├└]─ user:\[%d\] process .*
   [│ ]  [├└]─ pid:\[%d\] process .*$`,
			usernsid.Ino, pidnsid.Ino)))
	})

})
