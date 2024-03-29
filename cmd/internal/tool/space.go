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

// Space concatenates the passed strings, ensuring they are properly separated
// by only a single space between each source string, where necessary. Empty
// strings are correctly handled without inserting unnecessary separating
// spaces.
func Space(s string, more ...string) string {
	for _, s2 := range more {
		if s != "" && s2 != "" {
			s += " "
		}
		s += s2
	}
	return s
}
