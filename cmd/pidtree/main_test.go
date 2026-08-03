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
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("renders PID trees and branches", func() {

	var pidnsid species.NamespaceID
	var initpid, leafpid model.PIDType

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		var out safe.Buffer
		cmd := exec.Command("/bin/bash", "-c",
			`#!/bin/bash
# prints the namespace ID for the namespace type $1 of the current shell process.
echo -n "stage-1" > /proc/$$/comm
process_namespaceid () {
	LC_ALL=C stat -L /proc/$$/ns/$1 | awk '/Inode: [0-9]+/ { print $4 }'
}
export -f process_namespaceid
unshare -Upmfr /bin/bash -c "
	mount -t proc proc /proc
	echo -n "stage-2" > /proc/\$\$/comm
	process_namespaceid pid
	echo \$\$
	(echo \$BASHPID && read)
" && true
`)
		stdin, _ := cmd.StdinPipe() // need to keep an empty stdin so that the script stays
		sess := Successful(gexec.Start(cmd,
			io.MultiWriter(&out, GinkgoWriter), GinkgoWriter))
		DeferCleanup(func() {
			Expect(fmt.Fprint(stdin, "\n")).Error().To(Succeed())
			sess.Kill().Wait(2 * time.Second)
		})
		Eventually(out.String).Within(5 * time.Second).ProbeEvery(50 * time.Millisecond).
			Should(MatchRegexp(`^(\d+)\n(\d+)\n(\d+)\n$`))
		var pidnsino uint64
		Expect(fmt.Sscanf(out.String(), "%d\n%d\n%d\n",
			&pidnsino, &initpid, &leafpid)).Error().NotTo(HaveOccurred())
		pidnsid = species.NamespaceIDfromInode(pidnsino)
		Expect(pidnsid).NotTo(BeZero())
		Expect(initpid).To(Equal(model.PIDType(1)))
		Expect(leafpid).NotTo(Equal(model.PIDType(1)))
	})

	It("CLI w/o args renders PID tree", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)
		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`
(?m)^[│ ]+└─ "unshare" \(\d+\).*
[│ ]+└─ pid:\[%d\], owned by UID %d \(".*"\)
[│ ]+└─ "stage-2" \(\d+/1\).*
[│ ]+└─ "stage-2" \(\d+/%d\).*$`,
			pidnsid.Ino, os.Geteuid(), leafpid)))
	})

	DescribeTable("rejecting to render invalid PID branches",
		func(pid model.PIDType, pidnsid species.NamespaceID) {
			cmd := newRootCmd()
			cmd.SetArgs([]string{})
			var out safe.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			debug.SetWriter(cmd, GinkgoWriter)
			Expect(renderPIDBranch(cmd, &out, pid, pidnsid, nil)).To(HaveOccurred())
		},
		Entry(nil, model.PIDType(-1), species.NoneID),
		Entry(nil, model.PIDType(initpid), species.NamespaceIDfromInode(123)),
		Entry(nil, model.PIDType(-1), species.NamespaceIDfromInode(pidnsid.Ino)),
	)

	DescribeTable("CLI renders or rejects only a specific branch",
		func(nsa any, errmatcher types.GomegaMatcher, outmatchera any) {
			var ns string
			switch v := nsa.(type) {
			case string:
				ns = v
			case func() string:
				ns = v()
			default:
				panic(fmt.Sprintf("expected string or func() string, but got: %T", v))
			}

			var outmatcher types.GomegaMatcher
			switch v := outmatchera.(type) {
			case types.GomegaMatcher:
				outmatcher = v
			case func() types.GomegaMatcher:
				outmatcher = v()
			default:
				panic(fmt.Sprintf("expected types.GomegaMatcher or func() types.GomegaMatcher, but got: %T", v))
			}

			cmd := newRootCmd()
			cmd.SetArgs([]string{
				"--" + turtles.NoContainersFlagName,
				fmt.Sprintf("--pid=%d", initpid),
				fmt.Sprintf("--ns=%s", ns),
			})
			var out safe.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			debug.SetWriter(cmd, GinkgoWriter)

			err := cmd.Execute()
			Expect(err).To(errmatcher, "pid %d, ns %v", initpid, ns)
			Expect(out.String()).To(outmatcher)
		},
		Entry(nil,
			func() string { return fmt.Sprintf("%d", pidnsid.Ino) },
			Not(HaveOccurred()),
			func() types.GomegaMatcher {
				return MatchRegexp(fmt.Sprintf(`
(?m)^ +└─ pid:\[%d\], owned by UID %d \(".*"\)
\ +└─ "stage-2" \(\d+/1\).*
$`, pidnsid.Ino, os.Geteuid()))
			},
		),
		Entry(nil, "abc", HaveOccurred(), MatchRegexp(`Error: not a valid PID namespace ID`)),
		Entry(nil, "net:[12345]", HaveOccurred(), MatchRegexp(`Error: not a valid PID namespace ID:`)),
		Entry(nil, "pid:[12345]", HaveOccurred(), MatchRegexp(`Error: unknown PID namespace pid:`)),
	)

	It("CLI --foobar fails correctly", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--foobar"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("renders PIDs", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(`^pid:\[`))
	})

})
