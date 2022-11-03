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
	"io"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("captures stdout", func() {

	It("captures really looooong output", func() {
		By("redirecting the normal stdout and stderr into a buffer")
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		defer func() {
			os.Stdout = oldStdout
			os.Stderr = oldStderr
		}()
		r, w, err := os.Pipe()
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = r.Close()
			_ = w.Close()
		}()
		buff := gbytes.NewBuffer()
		go func() {
			_, _ = io.Copy(buff, r)
		}()
		os.Stdout = w
		os.Stderr = w

		By("testing that output before and after Stdouterr ends up in our buffer, and Stdouterr returns the output of its fn")
		s := strings.Repeat("Nobody expects the spanish inquisition! ", 10000)
		os.Stdout.WriteString("foobar!")
		out := Stdouterr(func() { os.Stdout.WriteString(s) })
		os.Stdout.WriteString("foobar!")
		Expect(out).To(Equal(s))
		out = Stdouterr(func() { os.Stderr.WriteString(s) })
		Expect(out).To(Equal(s))
		Expect(buff).To(gbytes.Say("foobar!foobar!"))
	})

})
