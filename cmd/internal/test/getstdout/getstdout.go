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

package getstdout

import (
	"bytes"
	"io"
	"os"
)

// Stdouterr runs function f and then returns the output from stdout and
// stderr as a string.
func Stdouterr(f func()) string {
	// Stash away os.Stdout so we can restore it when leaving this capture
	// function.
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()
	// Create a new OS pipe and ensure that we'll close both ends (file
	// descriptors) after we're done. Normally, this will do a double close to
	// the writing end, but as this doesn't matter, this keeps the code simple
	// yet we ensure to close under all circumstances, except for direct
	// meteor hits.
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer func() {
		r.Close()
		w.Close()
	}()
	// Kick off reading all data sent into the pipe off it so writers won't
	// stall indefinitely. In case of a non-EOF error, we append the pipe
	// reading failure reson to the captured output returned.
	stdout := make(chan string)
	go func() {
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, r); err != nil {
			stdout <- buf.String() + "\n" + err.Error()
			return
		}
		stdout <- buf.String()
	}()
	// All engines up and running, so we now redirect stdout and run the
	// specified f().
	os.Stdout = w
	os.Stderr = w
	f()
	// Done; signal the background pipe reading go routine to finish copying
	// and wait for it to return the captured stdout ... which we then return
	// to our caller.
	w.Close()
	return <-stdout
}
