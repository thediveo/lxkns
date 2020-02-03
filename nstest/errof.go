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

// Err returns only the last value for a function returning multiple values,
// and this value is considered to be the error return value.
func Err(v ...interface{}) (err error) {
	if len(v) < 2 {
		panic("function under test only returns a single value")
	}
	err, ok := v[len(v)-1].(error)
	if !ok {
		panic("function under test doesn't return an error as its last multi-value result")
	}
	return
}
