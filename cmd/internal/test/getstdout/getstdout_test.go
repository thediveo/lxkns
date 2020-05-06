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

package getstdout

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("captures stdout", func() {

	It("captures really looooong output", func() {
		s := strings.Repeat("Nobody expects the spanish inquisition! ", 10000)
		os.Stdout.WriteString("foobar!")
		out := Stdouterr(func() { os.Stdout.WriteString(s) })
		os.Stdout.WriteString("foobar!")
		Expect(out).To(Equal(s))
		out = Stdouterr(func() { os.Stderr.WriteString(s) })
		Expect(out).To(Equal(s))
	})

})
