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

package mountineer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/thediveo/lxkns/ops/mountineer/mntnssandbox" // ensure to pull in the pre-Go initializer.
)

// MntnsSandboxBinary is the name of a binary switching into a specified mount
// namespace and then going to sleep indefinitely. The rationale for a dedicated
// binary is that it will be much smaller than any binary containing all the
// nice lxkns discovery stuff and thus starts much faster and consumes less
// resources. A rough figure: ~840K on aarch64 without symbol and debug
// information.
const MntnsSandboxBinary = "mntnssandbox"

// sandboxBinary points either to a dedicated sandbox binary, or when
// this binary is unavailable, to our own binary as a fallback.
var sandboxBinary = "/proc/self/exe"
var separateSandboxBinary = false

// Is the dedicated mount namespace sandbox binary available? Then use that,
// otherwise fall back to our own binary.
func init() {
	pathname, err := exec.LookPath(MntnsSandboxBinary)
	if err != nil {
		return
	}
	pathname, err = filepath.Abs(pathname)
	if err == nil {
		sandboxBinary = pathname
		separateSandboxBinary = true
	}
}

// StandaloneSandboxBinary returns the pathname of a separate sandbox binary
// when found, or a zero string if the process needs to re-execute itself in
// order to start sandboxes.
func StandaloneSandboxBinary() string {
	if !separateSandboxBinary {
		return ""
	}
	return sandboxBinary
}

// NewPauseProcess starts a new pause process that immediately attaches itself
// to the mount namespace referenced by mntnsref. This function only returns
// after the pause process has finally attached to the mount namespace, or
// failed to do so. It does not return before success or failure is clear.
//
// Where available, NewPauseProcess will start the dedicated "mntnssandbox"
// binary. It will automatically fall back to re-executing itself for switching
// the mount namespace and pausing when the "mntnssandbox" binary cannot be
// found.
func NewPauseProcess(mntnsref string, usernsref string) (*exec.Cmd, error) {
	return newPauseProcess(sandboxBinary, mntnsref, usernsref)
}

// newPauseProcess starts a new pause process using the specified sandbox
// binary.
func newPauseProcess(binary string, mntnsref string, usernsref string) (*exec.Cmd, error) {
	sleepychild := exec.Command(binary)
	sleepychild.Env = append(os.Environ(),
		mntnssandbox.MntnsEnvironmentVariableName+"="+mntnsref)
	if usernsref != "" {
		sleepychild.Env = append(sleepychild.Env,
			mntnssandbox.UsernsEnvironmentVariableName+"="+usernsref)
	}
	// Stream the sandbox output live as we need to wait for the sandbox' "OK"
	// while it hasn't yet terminated.
	childout, err := sleepychild.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("failed to prepare to start pause process: %s", err.Error()))
	}
	// Collect all stderr output into a buffer so we have it ready when Wait()
	// returns and has closed the pipes.
	var childerr bytes.Buffer
	sleepychild.Stderr = &childerr
	// Take off...
	if err := sleepychild.Start(); err != nil {
		return nil, fmt.Errorf("cannot start pause process: %s", err.Error())
	}
	// Now wait for pause process to attach to the specified mount namespace;
	// this avoids race conditions where we otherwise would access the wrong
	// mount namespace. Of course, things might still go horribly wrong in after
	// starting the pause process, such as with invalid mount namespace
	// references, denied access, et cetera. So we need to deal with the pause
	// process permaturely terminating while we wait for the "OK" that will
	// never come...
	okch := make(chan bool)
	waiterrch := make(chan error, 1) // decouple sender from (maybe missing) consumer.
	go func() {
		r := bufio.NewReader(childout)
		s, err := r.ReadString('\n')
		if err != io.EOF {
			// Only signal if there is some output, but not if we got EOF
			// without any output because the pause process terminated
			// prematurely instead with an error.
			okch <- err == nil && s == "OK\n"
			close(okch)
		}
	}()
	// Wait for the sleepychild process to terminate in the background, even
	// after it has signalled "OK" and we have returned to the caller. The
	// caller thus should not call Wait() itself. In case we've already
	// returned, any error result will eventually be thrown away ... it doesn't
	// matter anymore. That's the reason why the channel is buffered in order to
	// avoid leaking blocked goroutines because there's none consuming the
	// channel messages anymores.
	go func() {
		waiterrch <- sleepychild.Wait()
		close(waiterrch)
	}()
	// Wait for the pause process to report that it has successfully attached to
	// the specified mount namespace, or alternatively that it has "crashed".
	select {
	case ok := <-okch:
		// If we got at least one line of pause process stdout output and if
		// that doesn't match what we would expect, then terminate the pause
		// process and report an error to the caller.
		if !ok {
			_ = sleepychild.Process.Kill()
			return nil, fmt.Errorf("error: unexpected pause process output")
		}
	case err := <-waiterrch:
		// The pause process has already terminated either with or without
		// error: that should not happen under normal circumstances, as the
		// pause process is supposed to pause indefinitely until we kill it.
		// Please note that we've already captured any stderr output in a buffer
		// because we cannot read the stderr pipe after Wait() has returned ...
		// which might have raced against us here.
		if childerrmsg := childerr.String(); childerrmsg != "" {
			return nil, fmt.Errorf("pause process failure: %s", childerrmsg)
		}
		if err != nil {
			return nil, err
		}
		return nil, errors.New("premature pause process termination")
	}
	// Keep the pause ... erm, running.
	return sleepychild, nil
}
