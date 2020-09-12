package gmodel

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// EqualNamespace returns a GomegaMatcher which compares an actual namespace to
// an expected namespace.
func EqualNamespace(expectedns interface{}) types.GomegaMatcher {
	return &equalNamespaceMatcher{expected: expectedns}
}

type equalNamespaceMatcher struct {
	expected interface{}
}

func (m *equalNamespaceMatcher) Match(actual interface{}) (bool, error) {
	if actual == m.expected {
		return true, nil
	}
	actualns, ok := actual.(model.Namespace)
	if !ok {
		return false, fmt.Errorf("SameNamespace expects a model.Namespace")
	}
	expectedns, ok := m.expected.(model.Namespace)
	if !ok {
		return false, fmt.Errorf("SameNamespace must be passed a model.Namespace")
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
			match = match && (actualns.Type() == species.CLONE_NEWUSER) &&
				actualns.(model.Ownership).UID() == expectedns.(model.Ownership).UID()
		}
	}
	return match, nil
}

func (m *equalNamespaceMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected namespace\n\t%s\nto match actual namespace\n\t%s",
		actual, m.expected)
}

func (m *equalNamespaceMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf(
		"Expected namespace\n\t%s\nto not match actual namespace\n\t%s",
		actual, m.expected)
}

// sameIDType returns true if both namespaces have the same ID and type, or if
// both are nil or referencing the same object.
func sameIDType(ns1, ns2 interface{}) bool {
	return ns1 == ns2 ||
		(ns1 != nil && ns2 != nil &&
			ns1.(model.Namespace).ID() == ns2.(model.Namespace).ID() &&
			ns1.(model.Namespace).Type() == ns2.(model.Namespace).Type())
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
			if child1 == child2 {
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
