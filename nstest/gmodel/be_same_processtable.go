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
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

// BeSameProcessTable returns a [types.GomegaMatcher] which compares an actual
// [model.ProcessTable] to an expected ProcessTable. This matcher doesn't test
// for deep equality (you could do this already using the existing matchers) but
// instead does only flat [model.Process] matching.
func BeSameProcessTable(expectedproctable interface{}) types.GomegaMatcher {
	return &beSameProcessTableMatcher{expected: expectedproctable}
}

type beSameProcessTableMatcher struct {
	expected interface{}
}

var dummyproctable = model.ProcessTable{}
var processtableT = reflect.TypeOf(dummyproctable)

func (matcher *beSameProcessTableMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil && matcher.expected == nil {
		return false, fmt.Errorf(
			// revive:disable-next-line:error-strings Gomega matchers
			// communicate useful messages, not Go platitudes.
			"Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead. This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}
	// "unpack" the ProcessTable-s
	actval := reflect.Indirect(reflect.ValueOf(actual))
	expval := reflect.Indirect(reflect.ValueOf(matcher.expected))
	if !(actval.IsValid() && expval.IsValid()) {
		return actval.IsValid() == expval.IsValid(), nil
	}
	if actval.Type() != processtableT {
		return false, fmt.Errorf(
			"BeSameProcessTable expects a model.ProcessTable, not a %T", actual)
	}
	if expval.Type() != processtableT {
		return false, fmt.Errorf(
			"BeSameProcessTable must be passed a model.ProcessTable, not a %T", matcher.expected)
	}
	actpt := actval.Interface().(model.ProcessTable)
	exppt := expval.Interface().(model.ProcessTable)
	if len(actpt) != len(exppt) {
		return false, nil
	}
	for pid, actproc := range actpt {
		expproc, ok := exppt[pid]
		if !ok || !similarProcess(actproc, expproc) {
			return false, nil
		}
	}
	return true, nil
}

func (matcher *beSameProcessTableMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected process table\n\t%+v\nto match actual process table\n\t%+v",
		actual, matcher.expected)
}

func (matcher *beSameProcessTableMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected process table\n\t%+v\nto not match actual process table\n\t%+v",
		actual, matcher.expected)
}
