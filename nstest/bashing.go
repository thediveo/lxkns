// Working with short auxiliary test shell scripts whose script sources can be
// kept together with the golang test code for better maintenance. Focuses on
// BASH shell scripts.

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
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo"
)

// defsfilename is the filename of a script file containing definitions for
// auxiliary Basher scripts, and these definitions then point to the temporary
// locations where these scripts have been written to.
const defsfilename = "basher-defs.sh"

// allowednamechars specifies the symbols allowed in shell environment and
// variable names.
var allowednamechars = regexp.MustCompile("[^A-Za-z0-9_]+")

// Basher manages (temporary) auxiliary scripts for setting up special
// environments for individual test cases. Make sure to always call Done at
// the end of a test case using Basher, preferably by immediately defer'ing
// the Done call.
//
// Scripts are added to a Basher by calling the Script method, giving the
// script a name and specifying the script itself. The script then gets
// written to a temporary file in a temporary directory, created especially
// for the test function which calls the Script method. When working with
// multiple test scripts, these scripts must be referenced via automatically
// injected environment variables named as the script (but without an ".sh"
// suffix).
//
// For example, script "a.sh" wants to call script "b.sh", so it needs to
// substitute "b.sh" by $b (or ${b}):
//   $b arg1 arg2 etc
//
// Invalid characters in shell script names, such as "-", will be replaced by
// "_" in the name of the corresponding environment variable.
type Basher struct {
	tmpdir   string            // temporary directory receiving scripts.
	defspath string            // path/filename to script with definitions, in temporary dir.
	scripts  map[string]string // maps script names to their temporary files.
}

// Done cleans up all temporary scripts and preferably is to be defer'ed by a
// test case immediately after creating a Basher.
func (b *Basher) Done() {
	if b.tmpdir != "" {
		// All we need to do is call remove all ;) This neatly removes the
		// temporary script directory with all its scripts.
		if err := os.RemoveAll(b.tmpdir); err != nil {
			panic(err.Error())
		}
		b.tmpdir = ""
	}
}

// Start starts the named script as a new TestCommand, with the given
// arguments.
func (b *Basher) Start(name string, args ...string) *TestCommand {
	name = strings.TrimSuffix(name, ".sh")
	scriptpath, ok := b.scripts[name]
	if !ok {
		panic(fmt.Sprintf("cannot run unknown script %q", name))
	}
	return NewTestCommand(scriptpath, args...)
}

// Script adds a (BASH) script with the given name. The script will
// automatically get the usual shebang prepended, so scripts should not
// include it themselves. An additional next line after the shebang will also
// be injected for sourcing a temporary file with script environment variables
// referencing the correct temporary script file path and filenames.
//
// For example: script "foo" (or "foo.sh") will have an associated environment
// variable "$foo" pointing to its temporary location. A script "foo-bar" has
// the associated environment variable "$foo_bar".
func (b *Basher) Script(name, script string) {
	b.init()
	b.addScript(name, script, false)
}

// Common adds an unnamed script with common definitions, which are then
// automatically made available to all (non-common) scripts.
func (b *Basher) Common(script string) {
	b.init()
	b.addScript(fmt.Sprintf("common%d", rand.Int()), script, true)
}

// addScript creates a temporary script file from the given script, and adds
// it to the known scripts as "name". If this is a "common" script, then it
// will automatically be sourced in all non-common scripts.
func (b *Basher) addScript(name, script string, common bool) {
	// Cut off any .sh suffix, if present. Then assign a full path to the
	// script, located in the temporary script directory.
	name = strings.TrimSuffix(name, ".sh")
	if _, ok := b.scripts[name]; ok {
		panic(fmt.Errorf("Basher: duplicate script name %q", name))
	}
	scriptpath := filepath.Join(b.tmpdir, name+".sh")
	b.scripts[name] = scriptpath
	// Set a new environment variable to the full path and script name
	// (including .sh) of the script. However, the name of the environment
	// variable itself will be sans any .sh suffix.
	envname := allowednamechars.ReplaceAllString(name, "_")
	f, err := os.OpenFile(b.defspath, os.O_APPEND|os.O_WRONLY, 0744)
	if err != nil {
		panic(fmt.Errorf(
			"Basher: cannot augment common definitions script %q, reason: %v",
			b.defspath, err))
	}
	defer f.Close()
	if !common {
		if _, err := f.WriteString(fmt.Sprintf("%s=%q\n", envname, scriptpath)); err != nil {
			panic(fmt.Errorf(
				"Basher: cannot augment common definitions script %q, reason: %v",
				b.defspath, err))
		}
	} else {
		if _, err := f.WriteString(fmt.Sprintf(". %q\n", scriptpath)); err != nil {
			panic(fmt.Errorf(
				"Basher: cannot augment common definitions script %q, reason: %v",
				b.defspath, err))
		}
	}
	// Create a new (executable) script file with the given name in the
	// temporary directory. We automatically prefix the script with a BASH
	// shebang and source the script alias definition so the script can be
	// called by their registered names but point to the correct temporary
	// location they were written to.
	f, err = os.OpenFile(scriptpath, os.O_WRONLY|os.O_CREATE, 0744)
	if err != nil {
		panic(fmt.Errorf(
			"Basher: cannot create temporary %q script as %q, reason: %v",
			name, scriptpath, err))
	}
	defer f.Close()
	scrpt := "#!/bin/bash\n"
	if !common {
		scrpt += ". " + b.defspath + "\n"
	}
	scrpt += script
	if _, err = f.WriteString(scrpt); err != nil {
		panic(fmt.Errorf(
			"Basher: cannot create temporary %q script as %q, reason: %v",
			name, scriptpath, err))
	}
}

// init initializes a Basher if it hasn't been initialized so far. Thus, init
// can be called multiple times without causing damage.
func (b *Basher) init() {
	if b.tmpdir != "" {
		return // already initialized, so we're done already.
	}
	// If this basher hasn't yet been initialized, we first create a temporary
	// directory with a prefix containing the current test's source code
	// filename (but without the ".go" file suffix) and a random suffix.
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
	// Set up a script file to be sourced by auxiliary scripts, which will
	// receive common environment variables definitions pointing to the
	// temporary locations of these aux scripts during a test.
	b.defspath = filepath.Join(b.tmpdir, defsfilename)
	f, err := os.OpenFile(b.defspath, os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		panic(fmt.Errorf(
			"Basher: failed to create %q for common definitions, reason: %v",
			b.defspath, err))
	}
	defer f.Close()
	if _, err = f.WriteString(fmt.Sprintf("#!/bin/bash\n")); err != nil {
		panic(fmt.Errorf(
			"Basher: cannot write %q with common definitions, reason: %v",
			b.defspath, err))
	}
}
