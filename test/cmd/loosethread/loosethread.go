// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Program loosethread creates a thread (task) that is attached to a newly
// created network namespace, without attaching this process' other tasks to the
// new network namespace.
//
// This program can be used when playing around with lxkns to ensure that loose
// threads are properly discovered and displayed in the web UI.
package main

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/thediveo/lxkns/ops"
	"golang.org/x/sys/unix"
)

func init() {
	runtime.LockOSThread()
}

func printns(nstype string) {
	netns := ops.NamespacePath("/proc/thread-self/ns/" + nstype)
	id, err := netns.ID()
	if err != nil {
		panic(fmt.Errorf("cannot determine network namespace of thread %d, reason: %w",
			unix.Gettid(), err))
	}
	fmt.Printf("task %d attached to %s:[%d]\n", unix.Gettid(), nstype, id.Ino)
}

func main() {
	go func() {
		runtime.LockOSThread()
		name := []byte("stray thread :p\x00")
		_ = unix.Prctl(unix.PR_SET_NAME, uintptr(unsafe.Pointer(&name[0])), 0, 0, 0)
		if err := unix.Unshare(unix.CLONE_NEWNET | unix.CLONE_NEWIPC); err != nil {
			panic(fmt.Errorf("cannot create new network namespace, reason: %w", err))
		}
		printns("ipc")
		printns("net")
		time.Sleep(time.Duration(1<<63 - 1))
	}()
	printns("ipc")
	printns("net")
	time.Sleep(time.Duration(1<<63 - 1))
}
