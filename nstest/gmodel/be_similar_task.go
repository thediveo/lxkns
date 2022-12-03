// Copyright 2023 Harald Albrecht.
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
	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
)

func BeSimilarTask(expectedtask any) types.GomegaMatcher {
	return &beSimilarTaskMatcher{expected: expectedtask}
}

type beSimilarTaskMatcher struct {
	expected any
}

var taskT = reflect.TypeOf(model.Task{})

func (matcher *beSimilarTaskMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil && matcher.expected == nil {
		return false, errors.New(
			// revive:disable-next-line:error-strings Gomega matchers
			// communicate useful messages, not Go platitudes.
			"Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead. This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}
	// "unpack" the Tasks
	actval := reflect.Indirect(reflect.ValueOf(actual))
	expval := reflect.Indirect(reflect.ValueOf(matcher.expected))
	if !(actval.IsValid() && expval.IsValid()) {
		return actval.IsValid() == expval.IsValid(), nil
	}
	if actval.Type() != taskT {
		return false, fmt.Errorf(
			"BeSimilarTask expects a model.Task, not a %T", actual)
	}
	if expval.Type() != taskT {
		return false, fmt.Errorf(
			"BeSimilarTask must be passed a model.Task, not a %T", matcher.expected)
	}
	acttask := actval.Interface().(model.Task)
	exptask := expval.Interface().(model.Task)
	if !similarTask(&acttask, &exptask) {
		return false, nil
	}
	return similarNamespacesSets(acttask.Namespaces, exptask.Namespaces), nil
}

func (matcher *beSimilarTaskMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected task\n\t%+v\nto match actual task\n\t%+v",
		actual, matcher.expected)
}

func (matcher *beSimilarTaskMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected task\n\t%+v\nto not match task process\n\t%+v",
		actual, matcher.expected)
}

func similarTask(task1, task2 *model.Task) bool {
	if task1 == nil || task2 == nil {
		return task1 == task2
	}
	matches, err := gomega.SatisfyAll(
		gomega.HaveField("TID", task2.TID),
		gomega.HaveField("ProTaskCommon.Name", task2.Name),
		gomega.HaveField("ProTaskCommon.Starttime", task2.Starttime),
	).Match(task1)
	if err != nil {
		panic(err)
	}
	return matches
}
