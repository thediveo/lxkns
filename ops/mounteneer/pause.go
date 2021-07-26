// Copyright 2021 Harald Albrecht.
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

package mounteneer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/thediveo/lxkns/mntnssandbox" // ensure to pull in the pre-Go initializer.
)

// MntnsSandboxBinary is the name of a binary switching into a specified mount
// namespace and then going to sleep indefinitely.
const MntnsSandboxBinary = "mntnssandbox"

// useFallback caches the outcome of whether the sandbox binary is available or
// we alternatively need to fallback to re-excuting our own (larger) binary
// instead of the small sandbox binary.
var useFallback = true

// NewPauseProcess starts a new pause process that immediately attaches itself
// to the mount namespace referenced by mntnsref. This function only returns
// after the pause process has finally attached to the mount namespace, or
// failed to do so. It does not return before success or failure is clear.
//
// Where available, NewPauseProcess will start the dedicated "mntnssandbox"
// binary. It will automatically fall back to re-executing itself for switching
// the mount namespace and pausing when the "mntnssandbox" binary cannot be
// found.
func NewPauseProcess(mntnsref string) (*exec.Cmd, error) {
	return reexecIntoPause(mntnsref)
}

// reexecIntoPause
func reexecIntoPause(mntnsref string) (*exec.Cmd, error) {
	sleepychild := exec.Command("/proc/self/exe")
	sleepychild.Env = append(os.Environ(),
		mntnssandbox.EnvironmentVariableName+"="+mntnsref)
	childout, err := sleepychild.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("cannot re-execute into pause: %s", err.Error()))
	}
	childerr, err := sleepychild.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("cannot re-execute into pause: %s", err.Error()))
	}
	if err := sleepychild.Start(); err != nil {
		panic(fmt.Sprintf("cannot re-execute into pause: %s", err.Error()))
	}
	// Wait for pause process to attach to the specified mount namespace; this
	// avoids race conditions where we otherwise would access the wrong mount
	// namespace.
	okch := make(chan bool)
	errch := make(chan error, 1) // decouple sender from (maybe missing) consumer.
	go func() {
		r := bufio.NewReader(childout)
		s, err := r.ReadString('\n')
		okch <- err == nil && s == "OK\n"
		close(okch)
	}()
	// Wait for the sleepychild process to terminate in the background, even
	// after it has signalled "OK" and we have returned to the caller. The
	// caller thus should not call Wait() itself. In case we've already
	// returned, any error result will eventually be thrown away ... it doesn't
	// matter anymore. That's the reason why the channel is buffered in order to
	// avoid leaking blocked goroutines because there's none consuming the
	// channel messages anymores.
	go func() {
		errch <- sleepychild.Wait()
		close(errch)
	}()
	// Wait for the pause process to report that it has successfully attached to
	// the specified mount namespace, or that it has "crashed".
	select {
	case ok := <-okch:
		// If we got at least one line of pause process stdout output and if
		// that doesn't match what we would expect, then terminate the pause
		// process and report an error to the caller.
		if !ok {
			sleepychild.Process.Kill()
			return nil, fmt.Errorf("error: unexpected pause process output")
		}
	case err := <-errch:
		// The pause process has already terminated either with or without
		// error: that should not happen under normal circumstances, as the
		// pause process is supposed to pause indefinitely until we kill it.
		// Please note that Wait() will correctly copy over all stdout/stderr
		// output before cleaning up, so we can safely catch here any stderr
		// output while Wait is running in a separate goroutine.
		if sleepyerrmsg, sleepyerr := ioutil.ReadAll(childerr); len(sleepyerrmsg) != 0 {
			return nil, fmt.Errorf("pause process failure: %s", string(sleepyerrmsg))
		} else {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("pause process failure: %s", sleepyerr.Error())
		}
	}
	// Keep the pause ... erm, running.
	return sleepychild, nil
}
