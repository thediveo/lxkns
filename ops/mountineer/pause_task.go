// Copyright 2023 Harald Albrecht.
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

package mountineer

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/thediveo/lxkns/model"
	"golang.org/x/sys/unix"
)

// newPauseTask starts a new pause task that immediately attaches itself
// to the mount namespace referenced by mntnsref. This function only returns
// after the pause task has finally attached to the mount namespace, or
// failed to do so. It does not return before success or failure is clear.
func newPauseTask(mntnsref string) (*pauseTask, error) {
	p := &pauseTask{
		done:    make(chan struct{}),
		outcome: make(chan any),
	}
	go p.pause(mntnsref)
	outcome := <-p.outcome
	if err, ok := outcome.(error); ok {
		return nil, err
	}
	p.tid = outcome.(model.PIDType)
	return p, nil
}

type pauseTask struct {
	tid       model.PIDType
	closeonce sync.Once
	done      chan struct{}
	outcome   chan any // either error or TID
}

// PID of the pauser task that can be used to access a mount namespace via the
// process filesystem. To be more precise, this actually is a TID, but then, who
// cares?
func (p *pauseTask) PID() model.PIDType {
	return p.tid
}

// Close the Pauser (by terminating it) and release allocated system resources.
// Close is idempotent.
func (p *pauseTask) Close() {
	p.closeonce.Do(func() { close(p.done) })
}

// pause must be run in a separate goroutine that then will lock itself to its
// underlying OS-level task (thread) and do its magic.
func (p *pauseTask) pause(mntnsref string) {
	runtime.LockOSThread() // but never unlock, as we'll taint the underlying task

	// revert sharing certain filesystem attributes between the process (other
	// tasks) with this task, where this sharing would otherwise prevent our
	// task from switching into different mount and optionally user namespaces.
	if err := unix.Unshare(unix.CLONE_FS); err != nil {
		p.outcome <- fmt.Errorf("pause task cannot unshare filesystem attributes, reason: %w", err)
		return
	}

	mntnsfd, err := unix.Open(mntnsref, unix.O_RDONLY, 0)
	if err != nil {
		p.outcome <- fmt.Errorf("invalid mount namespace reference, reason: %w", err)
		return
	}
	err = unix.Setns(mntnsfd, unix.CLONE_NEWNS)
	unix.Close(mntnsfd)
	if err != nil {
		p.outcome <- fmt.Errorf("cannot join mount namespace using reference %q, reason: %w",
			mntnsref, err)
		return
	}

	// We've successfully joined the namespace(s), so we have a useable TID to
	// address the procfs wormhole.
	p.outcome <- model.PIDType(unix.Gettid())

	// Now simply wait, keeping the wormhole open.
	<-p.done
}
