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

package matcher

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Handle fails in examples by printing out the Gomega failure message.
func stdoutFailures() {
	RegisterFailHandler(func(message string, _ ...int) {
		fmt.Println(message)
	})
}

// Ensure a fail handler for testable examples is installed.
func init() { stdoutFailures() }

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail) // handle failing Gomega tests correctly in Ginkgo context.
	RunSpecs(t, "lxkns/test/matcher package")
	stdoutFailures() // ensure the fail handler for testable examples get reinstalled.
}
