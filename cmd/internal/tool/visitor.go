// Copyright 2022 Harald Albrecht.
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

package tool

import (
	"fmt"
	"reflect"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
)

// SortRootNamespaces sorts the specified slice of model.Namespace objects by
// their IDs and then returns the sorted list as a slice consisting of
// reflection values for the individual namespace objects.
func SortRootNamespaces(roots reflect.Value) (children []reflect.Value) {
	namespaces, ok := roots.Interface().([]model.Namespace)
	if !ok {
		panic(fmt.Sprintf("expected []model.Namespace, got %T", roots))
	}
	namespaces = discover.SortNamespaces(namespaces)
	return ReflectValuesSlice(namespaces)
}

// ReflectValuesSlice returns a slice of reflect.Value objects for the specified
// slice of elements.
func ReflectValuesSlice[E any](e []E) []reflect.Value {
	count := len(e)
	slice := make([]reflect.Value, count)
	for idx := 0; idx < count; idx++ {
		slice[idx] = reflect.ValueOf(e[idx])
	}
	return slice
}
