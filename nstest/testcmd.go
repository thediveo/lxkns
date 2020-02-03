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

package nstest

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"
)

// BashExtractNamespaceID is bash script to extract the ID of a namespace from
// a namespace textual form "type:[id]".
const BashExtractNamespaceID = `sed -n -e 's/^.\+:\[\(.*\)\]/\1/p'`

// BashPrintNamespaceID returns bash script code for printing the ID of the
// namespace referenced by the specified path in the filesystem.
func BashPrintNamespaceID(path string) string {
	return fmt.Sprintf("readlink %s | %s", path, BashExtractNamespaceID)
}

// BashGetNamespaceRef returns a path to the namespace of type t of the
// current shell incarnation. It does basically what "$$" does, but as it
// evaluates only when it gets executed, it doesn't fall victim to early
// substitution.
func BashNamespacePath(t string) string {
	return `sed -n -e 's/^PPid:[[:space:]]\+\([[:digit:]]\+\)$/"\/proc\/\1\/ns\/` +
		t + `"/p' /proc/self/status`
}

// TestCommand is a command run as part of testing which, for instance, sets
// up some namespaces. The output of the TestCommand is streamed back in order
// to return data back to the test which is relevant to the test itself (such
// as namespace IDs). This return data is Decode'd as JSON, and it can be
// transferred in multiple and separate JSON data elements, to allow for a
// multi-stage test command (see also the Proceed method).
type TestCommand struct {
	cmd      *exec.Cmd      // the underlying OS command.
	childout io.ReadCloser  // command's stdout stream.
	childin  io.WriteCloser // command's stdin stream.
	dec      *json.Decoder  // JSON decoder for deserializing the command's stdout stream.
}

// NewTestCommand starts a command with arguments and then allows to read JSON
// data from the command and interact with the command in order to optionally
// step it through multiple stages under full control of a test. See the
// Decode and Proceed methods for details. When done, please Close a
// TestCommand.
func NewTestCommand(command string, args ...string) *TestCommand {
	cmd := &TestCommand{
		cmd: exec.Command(command, args...),
	}
	// Ensure that the test command and its children are in the same new
	// process group, so they can be stopped together.
	cmd.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Get the stdin and stdout streams for the soon-to-be child test command.
	childout, err := cmd.cmd.StdoutPipe()
	if err != nil {
		panic(err.Error())
	}
	cmd.childout = childout
	childin, err := cmd.cmd.StdinPipe()
	if err != nil {
		panic(err.Error())
	}
	cmd.childin = childin
	// And finally get a JSON decoder for decoding the test commands output
	// stream.
	cmd.dec = json.NewDecoder(childout)
	if err := cmd.cmd.Start(); err != nil {
		panic(err.Error())
	}
	return cmd
}

// Close completes the command by sending it an ENTER input and then closing
// the input pipe to the command. Then close waits at most 2s for the command
// to finish its business. If the command passes the timeout, then it will be
// killed hard.
func (cmd *TestCommand) Close() {
	cmd.Proceed()
	cmd.childin.Close()
	cmd.childout.Close()
	done := make(chan error)
	go func() { done <- cmd.cmd.Wait() }()
	select {
	case <-time.After(2 * time.Second):
		// And if thou'rt unwilling...
		cmd.cmd.Process.Kill()
	case <-done:
	}
}

// Decode reads JSON from the test command's output and tries to decode it
// into the data element specified.
func (cmd *TestCommand) Decode(v interface{}) {
	err := cmd.dec.Decode(v)
	if err != nil {
		panic(err)
	}
}

// Proceed sends the test command an ENTER input. This should be interpreted
// by the test command to advance into the next test phase for this command,
// or to finally terminate gracefully.
func (cmd *TestCommand) Proceed() {
	cmd.childin.Write([]byte{'\n'})
}
