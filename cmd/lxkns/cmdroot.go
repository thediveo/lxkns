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
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/caps"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"
	"golang.org/x/sys/unix"
)

// lxknsservice is the "root command" to be run after successfully parsing the
// CLI flags. We then here kick off the lxkns service itself.
func lxknsservice(cmd *cobra.Command, _ []string) error {
	if silent, _ := cmd.PersistentFlags().GetBool("silent"); silent {
		log.SetLevel(log.ErrorLevel)
	}
	if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("lxkns service debug logging enabled")
	}
	// initial cgroup hack around docker-compose not allowing for setting
	// "cgroupns: host" during deployment.
	switchedCgroup, _ := cmd.PersistentFlags().GetBool("cgroupswitched")
	switchCgroup, _ := cmd.PersistentFlags().GetBool("initialcgroup")
	if !switchedCgroup && switchCgroup {
		// Let's check if we're in a non-initial cgroup, such as when running
		// inside a Docker container on a "pure" unified cgroups v2 hierarchy
		// and with container cgroup'ing enabled...
		initialcgroupns := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP)
		initialcgroupnsid, ierr := initialcgroupns.ID()
		currentcgroupnsid, cerr := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		if ierr != nil || cerr != nil {
			log.Errorf("cannot determine initial and own cgroup namespaces, not switching cgroup namespace")
		} else if currentcgroupnsid != initialcgroupnsid {
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
			log.Infof("switching from current cgroup:[%d] into initial cgroup:[%d] and re-executing...",
				currentcgroupnsid.Ino, initialcgroupnsid.Ino)
			runtime.LockOSThread()
			if res, err := ops.Execute(func() interface{} {
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
				log.Debugf("current OS thread %sswitched to cgroup:[%d]", success, currentcgroupnsid.Ino)
				return unix.Exec(
					"/proc/self/exe",
					append([]string{os.Args[0], "--cgroupswitched"}, os.Args[1:]...),
					os.Environ(),
				)
			}, initialcgroupns); err != nil {
				log.Errorf("failed to switch to initial cgroup, err: %s", err.Error())
			} else {
				log.Errorf("failed to re-execute, err: %s", res)
				os.Exit(1)
			}
		}
	} else if switchedCgroup {
		// After re-execution log information about the current cgroup
		// namespace, which hopefully will be the initial cgroup namespace...
		initialcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/1/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		currentcgroupnsid, _ := ops.NewTypedNamespacePath("/proc/self/ns/cgroup", species.CLONE_NEWCGROUP).ID()
		isInitial := ""
		if currentcgroupnsid == initialcgroupnsid {
			isInitial = "initial "
		}
		log.Infof("re-executed in %scgroup:[%d]", isInitial, currentcgroupnsid.Ino)
		// Unfortunately, we end up here with /proc/self/stat stating our
		// process name as "exe", because we executed our own executable. This
		// is not terribly useful and user/admin friendly, so we try to set our
		// own process name from our first command line argument.
		runtime.LockOSThread() // this still runs on the main thread...!
		proc := model.NewProcess(model.PIDType(os.Getpid()))
		procname := append([]byte(proc.Basename()), 0)
		ptr := unsafe.Pointer(&procname[0])
		// prctl(PR_SET_NAME, ...) will silently truncate any process name
		// deemed too long, see also:
		// https://man7.org/linux/man-pages/man2/prctl.2.html
		if _, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
			log.Errorf("cannot fix process name: %s", syscall.Errno(errno).Error())
		} else {
			log.Debugf("fixed re-executed process name to %q", proc.Basename())
		}
		runtime.UnlockOSThread()
	}

	// And now for the real meat.
	log.Infof("This is the lxkns Linux-kernel namespaces discovery service and web app, version %s",
		lxkns.SemVersion)
	log.Infof("Copyright (c) Harald Albrecht, 2020..., see: https://github.com/thediveo/lxkns")
	log.Infof("This software is licensed under the Apache License, version 2.0, see: https://www.apache.org/licenses/LICENSE-2.0")

	log.Infof("running as user ID %d", os.Geteuid())
	mycaps := strings.Join(caps.ProcessCapabilities(model.PIDType(os.Getpid())), ", ")
	if mycaps == "" {
		mycaps = "<none>"
	}
	log.Infof("with effective capabilities: %s", mycaps)

	// Create the containerizer for the specified container engines...
	watchers := []watcher.Watcher{}
	if nodocker, _ := cmd.PersistentFlags().GetBool("nodocker"); !nodocker {
		apipath, err := cmd.PersistentFlags().GetString("docker")
		if err != nil {
			return err
		}
		log.Infof("synchronizing in background to Docker engine, API %s", apipath)
		watcher, err := moby.NewWatcher(apipath)
		if err != nil {
			return err
		}
		watchers = append(watchers, watcher)
	}
	if nocontainerd, _ := cmd.PersistentFlags().GetBool("nocontainerd"); !nocontainerd {
		apipath, err := cmd.PersistentFlags().GetString("containerd")
		if err != nil {
			return err
		}
		log.Infof("synchronizing in background to containerd engine, API %s", apipath)
		watcher, err := containerd.NewWatcher(apipath)
		if err != nil {
			return err
		}
		watchers = append(watchers, watcher)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cizer containerizer.Containerizer
	if len(watchers) > 0 {
		cizer = whalefriend.New(ctx, watchers)
		// Do not wait for the watchers to get online, they can do so in the
		// background.
		for _, w := range watchers {
			go func(w watcher.Watcher) {
				select {
				case <-w.Ready():
					idctx, _ := context.WithTimeout(ctx, 5*time.Second)
					log.Infof("synchronized to type %s container engine with ID %s at API %s",
						w.Type(), w.ID(idctx), w.API())
				case <-time.After(5 * time.Second):
					log.Warnf("container engine still offline for API %s ... silently trying", w.API())
				}
			}(w)
		}
	}

	// Fire up the service
	addr, _ := cmd.PersistentFlags().GetString("http")
	if _, err := startServer(addr, cizer); err != nil {
		log.Errorf("cannot start service, error: %s", err.Error())
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
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.BeforeCommand()
		},
		RunE: lxknsservice,
	}
	// Sets up the flags.
	pf := rootCmd.PersistentFlags()
	pf.Bool("debug", false, "enables debugging output")
	pf.Bool("silent", false, "silences everything below the error level")
	pf.String("http", "[::]:5010", "HTTP service address")
	pf.Duration("shutdown", 15*time.Second, "graceful shutdown duration limit")
	// Work around docker-compose currently having no means to set "cgroupns:
	// host" during deployment. There's a CLI flag, but no docker-composer
	// support, see also docker/compose issue #8167:
	// https://github.com/docker/compose/issues/8167.
	pf.Bool("initialcgroup", false, "switches into initial cgroup namespace")
	pf.Bool("cgroupswitched", false, "")
	_ = pf.MarkHidden("cgroupswitched")

	pf.Bool("nodocker", false, "disable Docker container discovery")
	pf.String("docker", "unix:///var/run/docker.sock", "Docker engine API socket address")
	pf.Bool("nocontainerd", false, "disable containerd container discovery")
	pf.String("containerd", "/run/containerd/containerd.sock", "containerd engine API socket address")

	cli.AddFlags(rootCmd)
	return
}
