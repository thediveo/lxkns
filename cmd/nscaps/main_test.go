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
	"github.com/thediveo/lxkns/ops"
)

var _ = Describe("renders branches", func() {

	It("CLI --foobar fails correctly", func() {
		os.Args = append(os.Args[:1], "--noengines", "--foobar")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("CLI rejects invalid target namespaces", func() {
		os.Args = append(os.Args[:1], "--noengines", "foo:[666]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: not a valid namespace:`))
	})

	It("CLI rejects invalid --ns", func() {
		os.Args = append(os.Args[:1],
			"--noengines",
			"--ns", "net:[666]",
			"net:[12345678]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: not a valid PID namespace:`))
	})

	It("CLI rejects valid --ns ID without --pid", func() {
		os.Args = append(os.Args[:1],
			"--noengines",
			"--ns", "666",
			"net:[12345678]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: --ns requires --pid`))
	})

	It("CLI rejects non-existing --ns ID", func() {
		os.Args = append(os.Args[:1],
			"--noengines",
			"--ns", "666",
			"--pid", "666",
			"net:[12345678]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown PID namespace`))
	})

	It("CLI rejects non-existing PID", func() {
		mypidns, err := ops.NamespacePath("/proc/self/ns/pid").ID()
		Expect(err).ToNot(HaveOccurred())
		os.Args = append(os.Args[:1],
			"--noengines",
			"--ns", fmt.Sprintf("%d", mypidns.Ino),
			"--pid", fmt.Sprintf("%d", ^uint32(0)),
			"net:[12345678]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown process PID .* in`))

		os.Args = append(os.Args[:1],
			"--noengines",
			"--pid", fmt.Sprintf("%d", ^uint32(0)),
			"net:[12345678]")
		out = getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown process PID .*`))
	})

	It("CLI rejects non-existing target namespace", func() {
		mypidns, err := ops.NamespacePath("/proc/self/ns/pid").ID()
		Expect(err).ToNot(HaveOccurred())
		os.Args = append(os.Args[:1],
			"--noengines",
			"--ns", fmt.Sprintf("%d", mypidns.Ino),
			"--pid", fmt.Sprintf("%d", os.Getpid()),
			"net:[12345678]")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown namespace net:`))
	})

	It("CLI w/o args fails", func() {
		os.Args = append(os.Args[:1], "--noengines")
		out := getstdout.Stdouterr(main)
		Expect(exitcode).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: expects 1 arg, received 0`))
	})

	It("CLI with target non-user namespace below process in owned user namespace", func() {
		if os.Geteuid() == 0 {
			Skip("only non-root")
		}
		mynetnsid, err := ops.NamespacePath("/proc/self/ns/net").ID()
		Expect(err).To(Succeed())
		os.Args = append(os.Args[:1],
			"--noengines",
			fmt.Sprintf("net:[%d]", mynetnsid.Ino))
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^⛛ user:\[%d\] process .*
├─ process .*
│     ⋄─ \(no effective capabilities\)
└─ target net:\[%d\] process .*
      ⋄─ \(no effective capabilities\)$`,
			initusernsid.Ino, mynetnsid.Ino)))
	})

	It("CLI with target non-user namespace below process in owned user namespace", func() {
		if os.Geteuid() == 0 {
			Skip("only non-root")
		}
		os.Args = append(os.Args[:1],
			"--noengines",
			fmt.Sprintf("net:[%d]", tnsid.Ino))
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^⛛ user:\[%d\] process .*
├─ process .*
│     ⋄─ \(no effective capabilities\)
└─ ✓ user:\[%d\] process .*
   └─ target net:\[%d\] process .*
         ⋄─ cap_audit_control .*$`,
			initusernsid.Ino, tuserid.Ino, tnsid.Ino)))
	})

	It("CLI with target non-user namespace at process", func() {
		os.Args = append(os.Args[:1],
			"--noengines",
			"-p", fmt.Sprintf("%d", tpid),
			fmt.Sprintf("net:[%d]", tnsid.Ino))
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^⛔ user:\[%d\] process .*
└─ ⛛ user:\[%d\] process .*
   ├─ process .*
   │     ⋄─ cap_audit_control .*
(   │     ⋄─ .*
)*   └─ target net:\[%d\] process .*
         ⋄─ cap_audit_control .*$`,
			initusernsid.Ino, tuserid.Ino, tnsid.Ino)))
	})

	It("CLI with process in other user namespace branch than target non-user namespace", func() {
		os.Args = append(os.Args[:1],
			"--noengines",
			"-p", fmt.Sprintf("%d", procpid),
			fmt.Sprintf("net:[%d]", tnsid.Ino))
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^⛔ user:\[%d\] process .*
├─ ⛛ user:\[%d\] process .*
│  └─ process .*
│        ⋄─ cap_audit_control .*
(│        ⋄─ .*
)*└─ ⛔ user:\[%d\] process .*
   └─ target net:\[%d\] process .*
         ⋄─ \(no capabilities\)$`,
			initusernsid.Ino, procusernsid.Ino, tuserid.Ino, tnsid.Ino)))
	})

})
