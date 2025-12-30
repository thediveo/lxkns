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
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/thediveo/lxkns/cmd/cli/style"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/spacer"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var (
	initialUsernsID               species.NamespaceID
	someProcPID                   model.PIDType
	someProcUsernsID              species.NamespaceID
	targetPID                     model.PIDType
	targetUsernsID, targetNetnsID species.NamespaceID
)

var allns *discover.Result

var _ = BeforeSuite(func() {
	initialUsernsID = species.NamespaceIDfromInode(spacetest.CurrentIno(unix.CLONE_NEWUSER))

	defer func(l *slog.Logger) { slog.SetDefault(l) }(slog.Default())
	slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

	By("spinning up a local spacer service")
	ctx, cancel := context.WithCancel(context.Background())
	spacerClient := spacer.New(ctx, spacer.WithErr(GinkgoWriter))
	DeferCleanup(func() {
		cancel()
		spacerClient.Close()
	})

	By("creating a child user namespace")
	subspaceClient, subspc := spacerClient.Subspace(true, false)
	DeferCleanup(func() {
		subspaceClient.Close()
	})
	someProcPID = model.PIDType(subspaceClient.PID())
	Expect(someProcPID).NotTo(Equal(model.PIDType(os.Getpid())))
	someProcUsernsID = species.NamespaceIDfromInode(spacetest.Ino(subspc.User, unix.CLONE_NEWUSER))

	By("creating another child user namespace and a network namespace")
	targetspaceClient, targetsubspc := spacerClient.Subspace(true, false)
	DeferCleanup(func() {
		targetspaceClient.Close()
	})
	targetPID = model.PIDType(targetspaceClient.PID())
	Expect(targetPID).NotTo(Equal(model.PIDType(os.Getpid())))
	Expect(targetPID).NotTo(Equal(someProcPID))
	targetUsernsID = species.NamespaceIDfromInode(spacetest.Ino(targetsubspc.User, unix.CLONE_NEWUSER))

	tnetns := targetspaceClient.NewTransient(unix.CLONE_NEWNET)
	targetNetnsID = species.NamespaceIDfromInode(spacetest.Ino(tnetns, unix.CLONE_NEWNET))

	allns = discover.Namespaces(discover.WithStandardDiscovery())
})

func TestNscapsCmd(t *testing.T) {
	style.PrepareForTest()

	format.MaxDepth = 4
	format.MaxLength = 10_000
	RegisterFailHandler(Fail)
	RunSpecs(t, "nscaps command")
}
