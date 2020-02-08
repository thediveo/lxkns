// Working with short auxiliary test shell scripts whose script sources can be
// kept together with the golang test code for better maintenance. Focuses on
// BASH shell scripts.

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

const NamespaceUtilsScript = `
# prints the namespace ID for the namespace referenced by path $1.
namespaceid () {
	readlink $1 | sed -n -e 's/^.\+:\[\(.*\)\]/\1/p'
}
# prints the namespace ID for the namespace type $1 of the current shell process.
process_namespaceid () {
	readlink /proc/$$/ns/$1 | sed -n -e 's/^.\+:\[\(.*\)\]/\1/p'
}
`
