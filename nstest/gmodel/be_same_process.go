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

package gmodel

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

// BeSameProcess returns a [types.GomegaMatcher] which compares an actual
// [model.Process] to an expected Process, using only a (semi-) flat comparism
// for equality. In particular, PID, PPID, Name, Cmdline, and Starttime are
// compared directly, since they are “flat” types. Additionally, it also checks
// the joined Namespaces for their IDs and types. But this matcher doesn't check
// the related Parent and Children Process objects; use BeSameTreeProcess
// instead.
func BeSameProcess(expectedprocess interface{}) types.GomegaMatcher {
	return &beSameProcessMatcher{expected: expectedprocess, intree: false}
}

// BeSameTreeProcess returns a [types.GomegaMatcher] which compares an actual
// [model.Process] to an expected Process, using only a (semi-) flat comparism
// for equality. In particular, PID, PPID, Name, Cmdline, and Starttime are
// compared directly, since they are “flat” types. The Parent and Children
// Process(es) are compared only for their PIDs (but not grandchildren, and so
// on). And the (joined) Namespaces are also only compared for their ID and
// type, but not anything beyond those two main properties.
func BeSameTreeProcess(expectedprocess interface{}) types.GomegaMatcher {
	return &beSameProcessMatcher{expected: expectedprocess, intree: true}
}

type beSameProcessMatcher struct {
	expected interface{}
	intree   bool // check Parent and Children?
}

var dummyproc = model.Process{}
var processT = reflect.TypeOf(dummyproc)

func (matcher *beSameProcessMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil && matcher.expected == nil {
		return false, errors.New(
			// revive:disable-next-line:error-strings Gomega matchers
			// communicate useful messages, not Go platitudes.
			"Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead. This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}
	// "unpack" the Process-es
	actval := reflect.Indirect(reflect.ValueOf(actual))
	expval := reflect.Indirect(reflect.ValueOf(matcher.expected))
	if !(actval.IsValid() && expval.IsValid()) {
		return actval.IsValid() == expval.IsValid(), nil
	}
	if actval.Type() != processT {
		return false, fmt.Errorf(
			"BeSame(Tree)Process expects a model.Process, not a %T", actual)
	}
	if expval.Type() != processT {
		return false, fmt.Errorf(
			"BeSame(Tree)Process must be passed a model.Process, not a %T", matcher.expected)
	}
	actproc := actval.Interface().(model.Process)
	expproc := expval.Interface().(model.Process)
	if match := similarProcess(&actproc, &expproc); !match || !matcher.intree {
		return match, nil
	}
	// The actual process should be similar to the expected process, and their
	// parents should be similar too. Please note that "similar" doesn't mean
	// deeply equal, but a limited "flat equality".
	if !similarProcess(&actproc, &expproc) ||
		!similarProcess(actproc.Parent, expproc.Parent) {
		return false, nil
	}
	// Children should all be similar ... let me rephrase that: each child of
	// the actual process should have a similar child of the expected process.
	for _, actchild := range actproc.Children {
		found := false
		for _, expchild := range expproc.Children {
			if found = similarProcess(actchild, expchild); found {
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	// Oh, and the joined namespaces should be similar, too.
	return similarNamespacesSets(actproc.Namespaces, expproc.Namespaces), nil
}

func (matcher *beSameProcessMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected process\n\t%+v\nto match actual process\n\t%+v",
		actual, matcher.expected)
}

func (matcher *beSameProcessMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected process\n\t%+v\nto not match actual process\n\t%+v",
		actual, matcher.expected)
}

// similarProcess does only a limited comparism of two processes, without
// checking references; that is, ignoring Parent, Children, and Namespaces.
func similarProcess(proc1, proc2 *model.Process) bool {
	if proc1 == nil || proc2 == nil {
		return proc1 == proc2
	}
	// "flat" equal: don't dig into the details ;)
	matches, err := gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"PID":       gomega.Equal(proc2.PID),
		"PPID":      gomega.Equal(proc2.PPID),
		"Name":      gomega.Equal(proc2.Name),
		"Cmdline":   gomega.Equal(proc2.Cmdline),
		"Starttime": gomega.Equal(proc2.Starttime),
	}).Match(*proc1)
	if err != nil {
		panic(err)
	}
	return matches
}
