// Copyright 2026 Harald Albrecht.
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

package main

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
)

// pruneNamespaces prunes namespaces that are no longer needed, because neither
// processes nor tasks are attached to them and they don't have a bind-mount
// reference.
func pruneNamespaces(d *discover.Result) {
	for _, namespaceMap := range d.Namespaces {
		// make sure to work on a copy of the keys as we're most probably going
		// to prune the map.
		for _, nsid := range slices.Collect(maps.Keys(namespaceMap)) {
			namespace, ok := namespaceMap[nsid]
			if !ok || isReferenced(namespace) {
				continue
			}
			// we will deal with PID and user namespaces only later, because we
			// need to prune the hierarchy unless there are still references
			// surviving.
			if _, ok := namespace.(model.Hierarchy); ok {
				continue
			}
			// disown this flat namespace immediately.
			if owner := namespace.Owner(); owner != nil {
				asUser(owner.(model.Namespace)).Disown(namespace)
			}
			delete(namespaceMap, nsid)
		}
	}
	// time to maybe say goodbye for those namespaces that try to still cling
	// on...
	//
	purgePIDNamespaces(d)
	purgeUserNamespaces(d)
}

// isReferenced returns true if the passed namespace isn't needed anymore,
// because it has neither processes nor tasks attached and it isn't bind-mounted
// either.
func isReferenced(namespace model.Namespace) bool {
	// as long as we're still having processes or even loose threads
	// attached, we won't let go...
	if len(namespace.Leaders()) != 0 || len(namespace.LooseThreads()) != 0 {
		return true
	}
	// is it bind-mounted...?
	ref := namespace.Ref()
	if len(ref) > 1 || (ref != nil && !strings.HasPrefix(ref[0], "/proc/")) {
		return true
	}
	return false // SAD, THIS IS SO VERY SAD!
}

// purgePIDNamespaces purges PID namespaces that aren't needed anymore,
// recursively from the bottom of the hierarchy purging. A PID namespace gets
// purged when there are no more processes or tasks attached AND the PID
// namespace has no child PID namespaces either.
func purgePIDNamespaces(d *discover.Result) {
	// First and depth-first, let go all those hierarchical PID namespaces
	// to which neither process nor task is attached anymore and that don't have
	// any child namespaces. We also disown those who have to go. Okay, the
	// metaphorical level slowly gets very creepy...
	for namespace := range allHierarchicals(d.Processes[1].Namespaces[model.PIDNS]) {
		// as we're iterating depth-first we only arrive here after we've gone
		// through all children (and grandchildren). If there are child PID
		// namespaces still left we cannot purge this PID namespace.
		if isReferenced(namespace.(model.Namespace)) ||
			len(namespace.Children()) != 0 {
			continue
		}
		// disown this PID namespace.
		if owner := namespace.(model.Namespace).Owner(); owner != nil {
			asUser(owner.(model.Namespace)).Disown(namespace.(model.Namespace))
		}
		// remove this PID namespace from its parent's list.
		if parent := namespace.Parent(); parent != nil {
			asHierarchical(parent.(model.Namespace)).RemoveChild(namespace)
		}
		delete(d.Namespaces[model.PIDNS], namespace.(model.Namespace).ID())
	}
}

func purgeUserNamespaces(d *discover.Result) {
	// Now for hierarchical user namespaces go also depth-first and remove those
	// user namespaces no longer needed.
	for _, namespace := range slices.Collect(allHierarchicals(d.Processes[1].Namespaces[model.UserNS])) {
		// as we're iterating depth-first we only arrive here after we've gone
		// through all children (and grandchildren). If there are child user
		// namespaces still left we cannot purge this user namespace.
		if isReferenced(namespace.(model.Namespace)) ||
			len(namespace.Children()) != 0 {
			continue
		}
		// disown this user namespace to in the end get really rid of it.
		if owner := namespace.(model.Namespace).Owner(); owner != nil {
			asUser(owner.(model.Namespace)).Disown(namespace.(model.Namespace))
		}
		if parent := namespace.Parent(); parent != nil {
			asHierarchical(parent.(model.Namespace)).RemoveChild(namespace)
		}
		delete(d.Namespaces[model.UserNS], namespace.(model.Namespace).ID())
	}
}

// asPlain returns the implementation for a "plain" namespace for the passed
// namespace interface value; otherwise, it panics.
func asPlain(namespace model.Namespace) *namespaces.PlainNamespace {
	if namespace == nil {
		return nil
	}
	var plain *namespaces.PlainNamespace
	switch namspc := namespace.(type) {
	case *namespaces.PlainNamespace:
		plain = namspc
	case *namespaces.HierarchicalNamespace:
		plain = &namspc.PlainNamespace
	case *namespaces.UserNamespace:
		plain = &namspc.PlainNamespace
	default:
		panic(fmt.Sprintf("cannot cast %T to *PlainNamespace", namespace))
	}
	return plain
}

// asHierarchical returns the implementation for a "hierarchical" PID or user
// namespace for the passed namespace interface value; otherwise, it panics.
func asHierarchical(namespace model.Namespace) *namespaces.HierarchicalNamespace {
	if namespace == nil {
		return nil
	}
	var hier *namespaces.HierarchicalNamespace
	switch namspc := namespace.(type) {
	case *namespaces.HierarchicalNamespace:
		hier = namspc
	case *namespaces.UserNamespace:
		hier = &namspc.HierarchicalNamespace
	default:
		panic(fmt.Sprintf("cannot cast %T to *HierarchicalNamespace", namespace))
	}
	return hier
}

// asHierarchical returns the implementation for a "user" namespace for the
// passed namespace interface value; otherwise, it panics.
func asUser(namespace model.Namespace) *namespaces.UserNamespace {
	if namespace == nil {
		return nil
	}
	var user *namespaces.UserNamespace
	switch namspc := namespace.(type) {
	case *namespaces.UserNamespace:
		user = namspc
	default:
		panic(fmt.Sprintf("cannot cast %T to *UserNamespace", namespace))
	}
	return user
}

// allHierarchical iterates over all namespaces in a tree of hierarchical PID or
// user namespaces, yielding the namespaces depth-first. It allows for
// modifications of the hierarchy at a specific node while still correctly
// iterating over the children of this namespace.
func allHierarchicals(root model.Namespace) iter.Seq[model.Hierarchy] {
	return func(yield func(model.Hierarchy) bool) {
		root, _ := root.(model.Hierarchy)
		_ = _yieldChildrenFirst(root, yield)
	}
}

// _yieldChildrenFirst calls the passed yield function for all children of the
// passed node (if any), and only then yields the node itself, return false in
// case the iteration should be aborted.
func _yieldChildrenFirst(node model.Hierarchy, yield func(model.Hierarchy) bool) bool {
	if node == nil {
		return true
	}
	for _, child := range slices.Clone(node.Children()) {
		if !_yieldChildrenFirst(child, yield) {
			return false
		}
	}
	return yield(node)
}
