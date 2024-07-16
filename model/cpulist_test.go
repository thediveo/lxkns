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

package model

import (
	"os"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("cpu affinity sets", func() {

	DescribeTable("parsing cpu sets",
		func(set CPUSet, expected CPUList) {
			Expect(cpuListFromSet(set)).To(Equal(expected))
		},
		Entry("nil set", nil, CPUList{}),
		Entry("all-zeros set", CPUSet{0}, CPUList{}),
		Entry("all-zeros set", CPUSet{0, 0}, CPUList{}),

		// all in first word
		Entry("single cpu #0", CPUSet{1 << 0, 0}, CPUList{{0, 0}}),
		Entry("single cpu #1", CPUSet{1 << 1}, CPUList{{1, 1}}),
		Entry("single cpu #63", CPUSet{1 << 63}, CPUList{{63, 63}}),
		Entry("single cpu #63, none else", CPUSet{1 << 63, 0, 0}, CPUList{{63, 63}}),
		Entry("cpus #1-3", CPUSet{0xe, 0}, CPUList{{1, 3}}),

		// skip first zero words
		Entry("single cpu #64", CPUSet{0, 1 << 0}, CPUList{{64, 64}}),

		// multiple cpu ranges in same word
		Entry("cpu #1-2, #62", CPUSet{1<<62 | 1<<2 | 1<<1}, CPUList{{1, 2}, {62, 62}}),

		// range across boundaries
		Entry("cpus #63-64", CPUSet{1 << 63, 1 << 0}, CPUList{{63, 64}}),
		Entry("cpus #63-127", CPUSet{1 << 63, ^uint64(0)}, CPUList{{63, 127}}),

		// multiple all-1s words
		Entry("cpu #0-127", CPUSet{^uint64(0), ^uint64(0)}, CPUList{{0, 127}}),

		// mixed
		Entry("cpu #0-64", CPUSet{^uint64(0), 1 << 0}, CPUList{{0, 64}}),
		Entry("cpu #0-64, 67", CPUSet{^uint64(0), 1<<3 | 1<<0}, CPUList{{0, 64}, {67, 67}}),
		Entry("cpu #65-127, 129", CPUSet{0, ^uint64(0) - 1, 1 << 1}, CPUList{{65, 127}, {129, 129}}),

		Entry("b/w", CPUSet{0xaa0}, CPUList{{5, 5}, {7, 7}, {9, 9}, {11, 11}}),
		Entry("art", CPUSet{0x5a0}, CPUList{{5, 5}, {7, 8}, {10, 10}}),
	)

	It("gets this process'es CPU affinity mask", func() {
		Expect(wordbytesize).To(Equal(uint64(64 /* bits in uint64 */ / 8 /* bits/byte*/)))
		cpulist := Successful(GetCPUList(PIDType(os.Getpid())))
		Expect(cpulist).NotTo(BeEmpty())
		Expect(setsize.Load()).NotTo(BeZero())
	})

})
