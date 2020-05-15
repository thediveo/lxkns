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
)

var _ = Describe("renders branches", func() {

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

	It("CLI w/o args fails", func() {
		oldExit := osExit
		defer func() { osExit = oldExit }()
		exit := 0
		osExit = func(code int) { exit = code }
		os.Args = os.Args[:1]
		out := getstdout.Stdouterr(main)
		Expect(exit).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: expects 1 arg, received 0`))
	})

	It("CLI with target non-user namespace below process", func() {
		os.Args = append(os.Args[:1], fmt.Sprintf("net:[%d]", tnsid.Ino))
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^⛛ user:\[%d\] process .*
├─ process .*
│     ⋄─ \(no capabilities\).*
└─ ✓ user:\[%d\] process .*
   └─ target net:\[%d\] process .*
         ⋄─ cap_audit_control .*$`,
			initusernsid.Ino, tuserid.Ino, tnsid.Ino)))
	})

	It("CLI with target non-user namespace at process", func() {
		os.Args = append(os.Args[:1], "-p", fmt.Sprintf("%d", tpid), fmt.Sprintf("net:[%d]", tnsid.Ino))
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
		os.Args = append(os.Args[:1], "-p", fmt.Sprintf("%d", procpid), fmt.Sprintf("net:[%d]", tnsid.Ino))
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
