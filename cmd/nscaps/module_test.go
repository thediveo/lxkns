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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var initusernsid species.NamespaceID

var procscripts = testbasher.Basher{}
var proccmd *testbasher.TestCommand
var procpid model.PIDType
var procusernsid species.NamespaceID

var targetscripts = testbasher.Basher{}
var targetcmd *testbasher.TestCommand
var tpid model.PIDType
var tuserid, tnsid species.NamespaceID

var allns *discover.Result

var _ = BeforeSuite(func() {
	procscripts.Common(nstest.NamespaceUtilsScript)
	procscripts.Script("main", `
process_namespaceid user
unshare -Ufr $stage2
`)
	procscripts.Script("stage2", `
echo $$
process_namespaceid user
read
`)
	proccmd = procscripts.Start("main")
	initusernsid = nstest.CmdDecodeNSId(proccmd)
	proccmd.Decode(&procpid)
	procusernsid = nstest.CmdDecodeNSId(proccmd)

	targetscripts.Common(nstest.NamespaceUtilsScript)
	targetscripts.Script("main", `
unshare -Unfr $stage2
`)
	targetscripts.Script("stage2", `
echo $$
process_namespaceid user
process_namespaceid net
read
`)
	targetcmd = targetscripts.Start("main")
	targetcmd.Decode(&tpid)
	tuserid = nstest.CmdDecodeNSId(targetcmd)
	tnsid = nstest.CmdDecodeNSId(targetcmd)

	allns = discover.Namespaces(discover.WithStandardDiscovery())
})

var _ = AfterSuite(func() {
	if proccmd != nil {
		proccmd.Close()
	}
	procscripts.Done()

	if targetcmd != nil {
		targetcmd.Close()
	}
	targetscripts.Done()
})

var oldexit func(int)
var exitcode int

var _ = BeforeEach(func() {
	oldexit = osExit
	exitcode = 0
	osExit = func(code int) { exitcode = code }
})

var _ = AfterEach(func() {
	osExit = oldexit
})

func TestNscapsCmd(t *testing.T) {
	style.PrepareForTest()
	RegisterFailHandler(Fail)
	RunSpecs(t, "nscaps command")
}
