// Copyright 2026 Harald Albrecht.
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

package cgrp

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/safe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("--cgroup flag", func() {

	It("rejects unknown cgroup name display modes", func() {
		cmd := cobra.Command{}
		cmd.SetArgs([]string{"--cgroup=foobar"})
		SetupCLI(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		Expect(cmd.Execute()).To(MatchError(MatchRegexp(`^invalid argument "foobar"`)))
	})

	It("does full/complete name display", func() {
		cmd := cobra.Command{}
		cmd.SetArgs([]string{"--cgroup=full"})
		SetupCLI(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		Expect(cmd.Execute()).To(Succeed())

		dm := CgroupDisplayName(&cmd)
		Expect(dm("/a/very/long/hex/0000111122223333444455556666777700001111222233334444555566667777/number")).To(Equal(
			"/a/very/long/hex/0000111122223333444455556666777700001111222233334444555566667777/number"))
	})

	It("shortens name display", func() {
		cmd := cobra.Command{}
		cmd.SetArgs([]string{"--cgroup=short"})
		SetupCLI(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		Expect(cmd.Execute()).To(Succeed())

		dm := CgroupDisplayName(&cmd)
		Expect(dm("/a/very/long/hex/0000111122223333444455556666777700001111222233334444555566667777/number")).To(Equal(
			"/a/very/long/hex/000011112222…/number"))
	})

})
