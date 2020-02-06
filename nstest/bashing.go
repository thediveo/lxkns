// Simplify working with slightly intricate and especially nested test helper
// external scripts, but keep the script sources near your test code for
// better maintenance.

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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo"
)

// Basher manages (temporary) helper script for setting up special environment
// for individual test cases. Make sure to always call Cleanup at the end of a
// test case using Basher, preferably by immediately defer'ing the Cleanup
// call.
type Basher struct {
	tmpdir    string            // temporary directory receiving/holding scripts.
	aliaspath string            // path/filename to alias script in temporary dir.
	scripts   map[string]string // maps script names to their temporary files.
}

// Cleanup cleans up all temporary scripts and preferably is to be defer'ed by
// a test case immediately after creating a Basher.
func (b *Basher) Cleanup() {
	if b.tmpdir != "" {
		if err := os.RemoveAll(b.tmpdir); err != nil {
			panic(err.Error())
		}
		b.tmpdir = ""
	}
}

// Start starts script as a new TestCommand.
func (b *Basher) Start(name string, args ...string) *TestCommand {
	name = strings.TrimSuffix(name, ".sh")
	scriptpath, ok := b.scripts[name]
	if !ok {
		panic(fmt.Sprintf("cannot run unknown script %q", name))
	}
	return NewTestCommand(scriptpath, args...)
}

// Script adds a (BASH) script with the given name. The script will
// automatically get the usual shebang added, if not present. The next line
// then will source a temporary file with script aliases referencing the
// correct temporary script files.
func (b *Basher) Script(name, script string) {
	// If this basher hasn't yet been initialized, we first create a temporary
	// directory with a prefix containing the current test's source code
	// filename (without the ".go" suffix) and a random suffix.
	if b.tmpdir == "" {
		currtest := ginkgo.CurrentGinkgoTestDescription()
		prefix := fmt.Sprintf("%s-line-%d-",
			strings.TrimSuffix(path.Base(currtest.FileName), ".go"),
			currtest.LineNumber)
		tmpdir, err := ioutil.TempDir("", prefix)
		if err != nil {
			panic(err.Error())
		}
		b.tmpdir = tmpdir
		b.scripts = make(map[string]string)
		b.aliaspath = filepath.Join(b.tmpdir, "basher-scripts.sh")
	}
	name = strings.TrimSuffix(name, ".sh")
	// Assign a full path to the script in the temporary script directory.
	if _, ok := b.scripts[name]; ok {
		panic(fmt.Sprintf("duplicate script name %q", name))
	}
	scriptpath := filepath.Join(b.tmpdir, name+".sh")
	b.scripts[name] = scriptpath
	// Add a shell variable.
	f, err := os.OpenFile(b.aliaspath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("#!/bin/bash\n%s=%q\n", name, scriptpath))
	if err != nil {
		panic(err.Error())
	}
	// Create a new (executable) script file with the given name in the
	// temporary directory. We automatically prefix the script with a BASH
	// shebang and source the script alias definition so the script can be
	// called by their registered names but point to the correct temporary
	// location they were written to.
	f, err = os.OpenFile(scriptpath, os.O_WRONLY|os.O_CREATE, 0744)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	_, err = f.WriteString("#!/bin/bash\n. " + b.aliaspath + "\n" + script)
	if err != nil {
		panic(err.Error())
	}
}
