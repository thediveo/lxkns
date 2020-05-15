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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rxtst "github.com/thediveo/gons/reexec/testing"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var initusernsid species.NamespaceID

var procscripts = testbasher.Basher{}
var proccmd *testbasher.TestCommand
var procpid lxkns.PIDType
var procusernsid species.NamespaceID

var targetscripts = testbasher.Basher{}
var targetcmd *testbasher.TestCommand
var tpid lxkns.PIDType
var tuserid, tnsid species.NamespaceID

var allns *lxkns.DiscoveryResult

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
	proccmd.Decode(&initusernsid)
	proccmd.Decode(&procpid)
	proccmd.Decode(&procusernsid)

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
	targetcmd.Decode(&tuserid)
	targetcmd.Decode(&tnsid)

	allns = lxkns.Discover(lxkns.FullDiscovery)
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

func TestMain(m *testing.M) {
	// Ensure that the registered handler is run in the re-executed child.
	// This won't trigger the handler while we're in the parent. We're using
	// gons' very special coverage profiling support for re-execution.
	mm := &rxtst.M{M: m}
	os.Exit(mm.Run())
}

func TestNscapsCmd(t *testing.T) {
	style.PrepareForTest()
	RegisterFailHandler(Fail)
	RunSpecs(t, "nscaps command")
}
