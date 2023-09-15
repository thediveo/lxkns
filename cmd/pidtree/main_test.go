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
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("renders PID trees and branches", func() {

	var pidnsid species.NamespaceID
	var initpid, leafpid model.PIDType

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		scripts := testbasher.Basher{}
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
		cmd := scripts.Start("main")
		pidnsid = nstest.CmdDecodeNSId(cmd)
		cmd.Decode(&initpid)
		Expect(initpid).To(Equal(model.PIDType(1)))
		cmd.Decode(&leafpid)

		DeferCleanup(func() {
			if cmd != nil {
				cmd.Close()
			}
			scripts.Done()
		})
	})

	It("CLI w/o args renders PID tree", func() {
		rootCmd := newRootCmd()
		out := bytes.Buffer{}
		rootCmd.SetOut(&out)
		rootCmd.SetArgs([]string{})
		Expect(rootCmd.Execute()).ToNot(HaveOccurred())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`
(?m)^[│ ]+└─ "unshare" \(\d+\).*
[│ ]+└─ pid:\[%d\], owned by UID %d \(".*"\)
[│ ]+└─ "stage2.sh" \(\d+/1\).*
[│ ]+└─ "stage2.sh" \(\d+/%d\).*$`,
			pidnsid.Ino, os.Geteuid(), leafpid)))
	})

	It("CLI renders only a branch", func() {
		out := bytes.Buffer{}
		cmd := &cobra.Command{}
		Expect(renderPIDBranch(cmd, &out, model.PIDType(-1), species.NoneID, nil)).To(HaveOccurred())
		Expect(renderPIDBranch(cmd, &out, model.PIDType(initpid), species.NamespaceIDfromInode(123), nil)).To(HaveOccurred())
		Expect(renderPIDBranch(cmd, &out, model.PIDType(-1), species.NamespaceIDfromInode(pidnsid.Ino), nil)).To(HaveOccurred())

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
\ +└─ "stage2.sh" \(\d+/1\).*
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
			rootCmd := newRootCmd()
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{
				fmt.Sprintf("--pid=%d", initpid),
				fmt.Sprintf("--ns=%s", run.ns),
			})
			err := rootCmd.Execute()
			Expect(err).To(run.m, "pid %d, ns %v", initpid, run.ns)
			Expect(out.String()).To(run.res)
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

		os.Args = []string{os.Args[0], "--foobar"}
		out := getstdout.Stdouterr(main)
		Expect(exit).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown flag: --foobar`))

		os.Args = []string{os.Args[0]}
		exit = 0
		out = getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(`^pid:\[`))
		Expect(exit).To(BeZero())
	})

})
