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

	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// BeSameNamespace returns a GomegaMatcher which compares an actual namespace
// to an expected namespace. A namespace is anything supporting at least the
// model.Namespace interface, as well as optionally (and depending on the type
// of namespace) model.Hierarchy and model.Ownership.
func BeSameNamespace(expectedns interface{}) types.GomegaMatcher {
	return &beSameNamespaceMatcher{expected: expectedns}
}

type beSameNamespaceMatcher struct {
	expected interface{}
}

func (matcher *beSameNamespaceMatcher) Match(actual interface{}) (bool, error) {
	if actual == matcher.expected {
		return true, nil
	}
	actualns, ok := actual.(model.Namespace)
	if !ok {
		return false, fmt.Errorf(
			"BeSameNamespace expects a model.Namespace, not a %T", actual)
	}
	expectedns, ok := matcher.expected.(model.Namespace)
	if !ok {
		return false, fmt.Errorf(
			"BeSameNamespace must be passed a model.Namespace, not a %T", matcher.expected)
	}
	match := sameIDType(actualns, expectedns) &&
		actualns.Ref() == expectedns.Ref() &&
		sameIDType(actualns.Owner(), expectedns.Owner()) &&
		sameLeaders(actualns.LeaderPIDs(), expectedns.LeaderPIDs())
	if match {
		switch actualns.Type() {
		case species.CLONE_NEWPID, species.CLONE_NEWUSER:
			match = sameIDType(
				actualns.(model.Hierarchy).Parent(),
				expectedns.(model.Hierarchy).Parent()) &&
				sameChildren(actualns.(model.Hierarchy).Children(),
					expectedns.(model.Hierarchy).Children())
			if match && actualns.Type() == species.CLONE_NEWUSER {
				match = actualns.(model.Ownership).UID() == expectedns.(model.Ownership).UID()
			}
		}
	}
	return match, nil
}

func (matcher *beSameNamespaceMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected namespace\n\t%s\nto match actual namespace\n\t%s",
		actual.(model.NamespaceStringer).String(), matcher.expected.(model.NamespaceStringer).String())
}

func (matcher *beSameNamespaceMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected namespace\n\t%s\nto not match actual namespace\n\t%s",
		actual, matcher.expected)
}

// sameIDType returns true if both namespaces have the same ID and type, or if
// both are nil or referencing the same object.
func sameIDType(ns1, ns2 interface{}) bool {
	return ns1 == ns2 ||
		(ns1 != nil && ns2 != nil &&
			ns1.(model.Namespace).ID() == ns2.(model.Namespace).ID() &&
			ns1.(model.Namespace).Type() == ns2.(model.Namespace).Type())
}

func similarNamespacesSets(nsset1, nsset2 model.NamespacesSet) bool {
	for idx := model.NamespaceTypeIndex(0); idx < model.NamespaceTypesCount; idx++ {
		if !sameIDType(nsset1[idx], nsset2[idx]) {
			return false
		}
	}
	return true
}

// sameLeaders returns true if both lists of PIDs contain the same PIDs.
func sameLeaders(l1, l2 []model.PIDType) bool {
	if len(l1) != len(l2) {
		return false
	}
	for _, pid1 := range l1 {
		found := false
		for _, pid2 := range l2 {
			if pid1 == pid2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// sameChildren returns true if both lists of children contain the same
// children.
func sameChildren(l1, l2 []model.Hierarchy) bool {
	if len(l1) != len(l2) {
		return false
	}
	for _, child1 := range l1 {
		found := false
		for _, child2 := range l2 {
			if sameIDType(child1, child2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
