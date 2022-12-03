// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package discover

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("procfs references", func() {

	DescribeTable("parsing PIDs in procfs paths",
		func(path string, expectedpid int) {
			Expect(PIDfromPath(path)).To(Equal(model.PIDType(expectedpid)))
		},
		EntryDescription("%s -> %d"),
		Entry(nil, "/proc/1234/foo", 1234),
		Entry(nil, "/proc/1234", 1234),
		Entry(nil, "/proc/-1234", 0),
		Entry(nil, "/proc/foobar/", 0),
		Entry(nil, "/proc/foobar", 0),
		Entry(nil, "/proc", 0),
		Entry(nil, "/proc/", 0),
		Entry(nil, "/", 0),
		Entry(nil, "proc/1234/", 0),
		Entry(nil, "/abc/def", 0),
	)

	procs := model.ProcessTable{
		42:  {ProTaskCommon: model.ProTaskCommon{Starttime: 100}},
		666: {ProTaskCommon: model.ProTaskCommon{Starttime: 666}},
	}

	DescribeTable("preferring procfs references",
		func(newly, known model.NamespaceRef, expected bool) {
			Expect(NewlyProcfsPathIsBetter(newly, known, procs)).To(
				Equal(expected))
		},
		func(newly, known model.NamespaceRef, expected bool) string {
			return fmt.Sprintf("new %s versus known %s -> %t", newly.String(), known.String(), expected)
		},
		Entry(nil, r("/proc/42", "/foobar"), r("/proc/666", "/foobar"), true),
		Entry(nil, r("/proc/666", "/foobar"), r("/proc/42", "/foobar"), false),
		Entry(nil, r("/proc/42", "/foobar"), r("/proc/666", "/barz"), false),

		Entry(nil, r(), r(), false),
		Entry(nil, r(), r("/foo"), false),
		Entry(nil, r("/proc/123"), r("/proc/123"), false),

		Entry(nil, r("/proc/42", "/abc"), r("/abc/def", "/abc"), false),
		Entry(nil, r("/abc/def", "/abc"), r("/proc/42", "/abc"), false),
		Entry(nil, r("/proc/42", "/abc"), r("/proc/42", "/abc"), false),

		Entry(nil, r("/proc/42", "/abc"), r("/proc/777", "/abc"), false),
		Entry(nil, r("/proc/777", "/abc"), r("/proc/42", "/abc"), false),
	)

})

func r(elements ...string) (ref model.NamespaceRef) {
	ref = append(ref, elements...)
	return
}
