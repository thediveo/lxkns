// Copyright 2024 Harald Albrecht.
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

//go:build go1.23

package model

import (
	"iter"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

func taskNames(it iter.Seq[*Task]) []string {
	names := []string{}
	for task := range it {
		names = append(names, task.Name)
	}
	return names
}

// Returns only the first task name and then tells the iterator to shut up. This
// exercises also the unhappy little paths...
func firstTaskName(it iter.Seq[*Task]) []string {
	names := []string{}
	for task := range it {
		names = append(names, task.Name)
		break
	}
	return names
}

var _ = Describe("process and task iterators", func() {

	pt := ProcessTable{
		42: {
			ProTaskCommon: ProTaskCommon{Name: "pfoo"},
			Tasks: []*Task{
				{
					ProTaskCommon: ProTaskCommon{Name: "rfoo"},
				},
				{
					ProTaskCommon: ProTaskCommon{Name: "rbar"},
				},
				{
					ProTaskCommon: ProTaskCommon{Name: "rbaz"},
				},
			},
			Children: []*Process{
				{
					ProTaskCommon: ProTaskCommon{Name: "pfrobz"},
					Tasks: []*Task{
						{
							ProTaskCommon: ProTaskCommon{Name: "ffrobz"},
						},
					},
				},
				{
					ProTaskCommon: ProTaskCommon{Name: "pgnampf"},
					Tasks: []*Task{
						{
							ProTaskCommon: ProTaskCommon{Name: "fgnampf"},
						},
					},
				},
			},
		},
	}

	It("iterates over tasks of a single process", func() {
		Expect(taskNames(Tasks(pt, 0))).To(BeEmpty())
		Expect(firstTaskName(Tasks(pt, 42))).To(ConsistOf("rfoo"))
		Expect(taskNames(Tasks(pt, 42))).To(ConsistOf(
			"rfoo", "rbar", "rbaz"))
	})

	It("iterates over tasks of processes recursively", func() {
		Expect(taskNames(TasksRecursive(pt, 0))).To(BeEmpty())
		Expect(firstTaskName(TasksRecursive(pt, 42))).To(ConsistOf("rfoo"))
		Expect(taskNames(TasksRecursive(pt, 42))).To(ConsistOf(
			"rfoo", "rbar", "rbaz", "ffrobz", "fgnampf"))
	})

})
