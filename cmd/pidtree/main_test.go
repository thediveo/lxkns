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
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("renders PID trees and branches", func() {

	var scripts testbasher.Basher
	var cmd *testbasher.TestCommand
	var pidnsid species.NamespaceID
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

	It("CLI w/o args renders PID tree", func() {
		defer func() { rootCmd.SetOut(nil) }()
		out := bytes.Buffer{}
		rootCmd.SetOut(&out)
		rootCmd.SetArgs([]string{})
		Expect(rootCmd.Execute()).ToNot(HaveOccurred())
		tree := out.String()
		Expect(tree).To(MatchRegexp(fmt.Sprintf(`
(?m)^[│ ]+└─ "unshare" \(\d+\)
[│ ]+└─ pid:\[%d\], owned by UID %d \(".*"\)
[│ ]+└─ "stage2.sh" \(\d+/1\)
[│ ]+└─ "stage2.sh" \(\d+/%d\)$`,
			pidnsid.Ino, os.Geteuid(), leafpid)))
	})

	It("CLI renders only a branch", func() {
		out := bytes.Buffer{}
		Expect(renderPIDBranch(&out, lxkns.PIDType(-1), species.NoneID)).To(HaveOccurred())
		Expect(renderPIDBranch(&out, lxkns.PIDType(initpid), species.NamespaceIDfromInode(123))).To(HaveOccurred())
		Expect(renderPIDBranch(&out, lxkns.PIDType(-1), species.NamespaceIDfromInode(pidnsid.Ino))).To(HaveOccurred())

		defer func() {
			rootCmd.SetOut(nil)
			_ = rootCmd.PersistentFlags().Set("pid", "0")
			_ = rootCmd.PersistentFlags().Set("ns", "")
		}()
		for _, run := range []struct {
			ns  string
			m   OmegaMatcher
			res OmegaMatcher
		}{
			{
				ns: fmt.Sprintf("%d", pidnsid.Ino),
				m:  Not(HaveOccurred()),
				res: MatchRegexp(fmt.Sprintf(`
(?m)^ +└─ pid:\[%d\], owned by UID %d \(".*"\)
\ +└─ "stage2.sh" \(\d+/1\)
$`,
					pidnsid.Ino, os.Geteuid())),
			},
			{
				ns:  "abc",
				m:   HaveOccurred(),
				res: MatchRegexp(`Error: not a valid PID namespace ID`),
			},
			{
				ns:  "net:[12345]",
				m:   HaveOccurred(),
				res: MatchRegexp(`Error: not a valid PID namespace ID:`),
			},
			{
				ns:  "pid:[12345]",
				m:   HaveOccurred(),
				res: MatchRegexp(`Error: unknown PID namespace pid:`),
			},
		} {
			out.Reset()
			rootCmd.SetOut(&out)
			rootCmd.SetArgs([]string{
				fmt.Sprintf("--pid=%d", initpid),
				fmt.Sprintf("--ns=%s", run.ns),
			})
			err := rootCmd.Execute()
			Expect(err).To(run.m, "pid %d, ns %v", initpid, run.ns)
			tree := out.String()
			Expect(tree).To(run.res)
		}
	})

	It("runs and fails correctly", func() {
		oldArgs := os.Args
		oldExit := osExit
		defer func() {
			osExit = oldExit
			os.Args = oldArgs
		}()
		exit := 0
		osExit = func(code int) { exit = code }

		defer func() {
			rootCmd.SetOut(nil)
			_ = rootCmd.PersistentFlags().Set("pid", "0")
			_ = rootCmd.PersistentFlags().Set("ns", "")
		}()

		out := bytes.Buffer{}
		rootCmd.SetOut(&out)
		rootCmd.SetArgs(nil)
		os.Args = []string{os.Args[0], "--foobar"}
		main()
		Expect(exit).To(Equal(1))
		Expect(out.String()).To(MatchRegexp(`^Error: unknown flag: --foobar`))

		out.Reset()
		rootCmd.SetOut(&out)
		rootCmd.SetArgs(nil)
		os.Args = os.Args[:1]
		exit = 0
		main()
		Expect(out.String()).To(MatchRegexp(`^pid:\[`))
		Expect(exit).To(BeZero())
	})

})
