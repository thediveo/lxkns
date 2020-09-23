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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("renders user namespaces", func() {

	var scripts testbasher.Basher
	var cmd *testbasher.TestCommand
	var initusernsid, usernsid, netnsid species.NamespaceID

	BeforeEach(func() {
		cmd = nil
		scripts = testbasher.Basher{}
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
process_namespaceid user
unshare -Unfr $stage2
`)
		scripts.Script("stage2", `
process_namespaceid user
process_namespaceid net
read
`)
		cmd = scripts.Start("main")
		initusernsid = nstest.CmdDecodeNSId(cmd)
		usernsid = nstest.CmdDecodeNSId(cmd)
		netnsid = nstest.CmdDecodeNSId(cmd)
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
		os.Args = append(os.Args[:1], "--foobar")
		out := getstdout.Stdouterr(main)
		Expect(exit).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("CLI w/o args renders user tree", func() {
		os.Args = os.Args[:1]
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^[├└]─ user:\[%d\] .*$`,
			usernsid.Ino)))
	})

	It("CLI -d renders user tree with owned namespaces", func() {
		os.Args = append(os.Args[:1], "-d")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ net:\[%d\] process .*$`,
			usernsid.Ino, netnsid.Ino)))
	})

	It("CLI -f filters owned namespaces", func() {
		os.Args = append(os.Args[:1], "-d", "-f=pid")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).ToNot(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ .*$`,
			usernsid.Ino)))

		os.Args = append(os.Args[:1], "-d", "-f=ipc,net,pid")
		out = getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ net:\[%d\] process .*$`,
			usernsid.Ino, netnsid.Ino)))
	})

})
