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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/log"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/cmd/internal/caps"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// reexec reexecutes itself in the initial cgroup namespace, if necessary, and
// only then terminates this process. In case of re-execution errors, it logs
// these errors and then exits, too. If switching wasn't necessary, it simply
// returns.
func reexec() {
	// Let's check if we're in a non-initial cgroup, such as when running
	// inside a Docker container on a "pure" unified cgroups v2 hierarchy
	// and with container cgroup'ing enabled...
	initialcgroupns := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP)
	initialcgroupnsid, ierr := initialcgroupns.ID()
	currentcgroupnsid, cerr := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
	if ierr != nil || cerr != nil {
		slog.Error("cannot determine initial and own cgroup namespaces, not switching cgroup namespace")
		return
	}
	slog.Debug("cgroup namespaces",
		slog.Uint64("current", currentcgroupnsid.Ino), slog.Uint64("initial", initialcgroupnsid.Ino))
	if currentcgroupnsid != initialcgroupnsid {
		// In order to safely switch the cgroup namespace in a Golang app
		// with potentially several OS threads bouncing around by now we can
		// only lock our current OS thread, switch into the initial cgroup
		// namespace, and finally reexecute ourselves again. We don't use
		// our re-execution support package here, as that on purpose is
		// designed to not be re-executable from a child and provides a
		// JSON-oriented result interface (which we don't need). Instead, we
		// just use the few and simple primitives, namely
		// runtime.LockOSThread() and ops.Execute(). Everything else is
		// debug logging and error handling.
		slog.Info("switching from current cgroup into initial cgroup and re-executing...",
			slog.Uint64("current", currentcgroupnsid.Ino), slog.Uint64("initial", initialcgroupnsid.Ino))
		runtime.LockOSThread()
		if res, err := ops.Execute(func() error {
			// tee hee, while the process might still be in its original,
			// but not initial, cgroup namespace, this particular OS-level
			// task/thread should now be in the initial cgroup namespace. So
			// we need to query the current cgroup namespace by TID, not
			// PID. Please note that we don't need to use
			// "/proc/self/task/$TID/ns/cgroup", as all tasks are directly
			// accessible at the /proc/$TID level; they're just not listed,
			// yet still there.
			currentcgroupnsid, _ := ops.NewTypedNamespacePath(
				fmt.Sprintf("/proc/%d/ns/cgroup", syscall.Gettid()),
				species.CLONE_NEWCGROUP).ID()
			success := ""
			if currentcgroupnsid == initialcgroupnsid {
				success = "successfully "
			}
			slog.Debug(fmt.Sprintf("(parent) current OS thread %sswitched to cgroup", success),
				slog.Uint64("current", currentcgroupnsid.Ino))
			return unix.Exec(
				"/proc/self/exe",
				append([]string{os.Args[0], "--cgroupswitched"}, os.Args[1:]...),
				os.Environ(),
			)
		}, initialcgroupns); err != nil {
			slog.Error("failed to switch to initial cgroup", slog.String("err", err.Error()))
		} else {
			slog.Error("failed to re-execute", slog.String("err", res.Error()))
			os.Exit(1)
		}
	}
}

// reexeced runs in the re-executed child process and logs information about the
// current cgroup namespace, which hopefully will be the initial cgroup
// namespace. It fixes the process name to show the binary's base name, as
// opposed to "exe" (which initially set due to re-execution /proc/self/exe).
func reexeced() {
	initialcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP).ID()
	currentcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
	slog.Info("re-executed",
		slog.Bool("initial", currentcgroupnsid == initialcgroupnsid),
		slog.Int64("cgroup", int64(currentcgroupnsid.Ino)))
	// Unfortunately, we end up here with /proc/self/stat stating our
	// process name as "exe", because we executed our own executable. This
	// is not terribly useful and user/admin friendly, so we try to set our
	// own process name from our first command line argument.
	runtime.LockOSThread() // this still runs on the main thread...!
	proc := model.NewProcess(model.PIDType(os.Getpid()), false)
	procname := append([]byte(proc.Basename()), 0)
	ptr := unsafe.Pointer(&procname[0]) // #nosec G103
	// prctl(PR_SET_NAME, ...) will silently truncate any process name
	// deemed too long, see also:
	// https://man7.org/linux/man-pages/man2/prctl.2.html
	if _, _, errno := syscall.RawSyscall6(
		syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
		slog.Error("cannot fix process name",
			slog.String("err", syscall.Errno(errno).Error()))
	} else {
		slog.Debug("fixed re-executed process name",
			slog.String("name", proc.Basename()))
	}
	runtime.UnlockOSThread()
}

// lxknsservice is the "root command" to be run after successfully parsing the
// CLI flags. We then here kick off the lxkns service itself.
func lxknsservice(cmd *cobra.Command, _ []string) error {
	// initial cgroup hack around docker-compose not allowing for setting
	// "cgroupns: host" during deployment.
	switchedCgroup, _ := cmd.PersistentFlags().GetBool("cgroupswitched")
	switchCgroup, _ := cmd.PersistentFlags().GetBool("initialcgroup")
	if !switchedCgroup && switchCgroup {
		reexec()
	} else if switchedCgroup {
		reexeced()
	}

	// And now for the real meat.
	slog.Info("This is the lxkns Linux-kernel namespaces discovery service and web app",
		slog.String("version", lxkns.SemVersion))
	slog.Info("Copyright (c) Harald Albrecht, 2020, 2025, ...; see: https://github.com/thediveo/lxkns")
	slog.Info("This software is licensed under the Apache License, version 2.0, see: https://www.apache.org/licenses/LICENSE-2.0")

	if pausebin := mountineer.StandaloneSandboxBinary(); pausebin != "" {
		slog.Info("using optimized pandora's sandbox binary", slog.String("path", pausebin))
	}

	slog.Info("running as", slog.Int("userid", os.Geteuid()))
	mycaps := strings.Join(caps.ProcessCapabilities(model.PIDType(os.Getpid())), ", ")
	if mycaps == "" {
		mycaps = "<none>"
	}
	slog.Info("with effective capabilities", slog.String("effcaps", mycaps))

	slog.Info("available decorator plugins",
		slog.String("decorators",
			strings.Join(plugger.Group[decorator.Decorate]().Plugins(), ",")))

	// Create the containerizer for the specified container engines...
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()

	// Fire up the service
	addr, _ := cmd.PersistentFlags().GetString("http")
	if _, err := startServer(addr, cizer); err != nil {
		slog.Error("cannot start service", slog.String("err", err.Error()))
		os.Exit(1)
	}
	stopit := make(chan os.Signal, 1)
	signal.Notify(stopit, syscall.SIGINT)
	signal.Notify(stopit, syscall.SIGTERM)
	signal.Notify(stopit, syscall.SIGQUIT)
	<-stopit
	maxwait, _ := cmd.PersistentFlags().GetDuration("shutdown")
	stopServer(maxwait)
	return nil
}

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lxkns",
		Short:   "lxkns provides Linux-kernel namespace discovery as a service",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: lxknsservice,
	}
	// Sets up the flags.
	pf := rootCmd.PersistentFlags()
	pf.String("http", "[::]:5010", "HTTP service address")
	pf.Duration("shutdown", 15*time.Second, "graceful shutdown duration limit")
	// Work around docker-compose currently having no means to set "cgroupns:
	// host" during deployment. There's a CLI flag, but no docker-composer
	// support, see also docker/compose issue #8167:
	// https://github.com/docker/compose/issues/8167.
	pf.Bool("initialcgroup", false, "switches into initial cgroup namespace")
	pf.Bool("cgroupswitched", false, "")
	_ = pf.MarkHidden("cgroupswitched")

	clippy.AddFlags(rootCmd)
	return
}
