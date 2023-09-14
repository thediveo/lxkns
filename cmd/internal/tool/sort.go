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

import "golang.org/x/exp/slices"

// Sort sorts a copy of the passed slice using the specified less function and
// returns the sorted copy.
func Sort[E any](s []E, less func(e1, e2 E) int) []E {
	s = slices.Clone(s)
	slices.SortFunc(s, less)
	return s
}
