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

package nstypes

import (
	"fmt"
)

func ExampleNamespaceType_String() {
	for _, t := range []NamespaceType{
		CLONE_NEWNS, CLONE_NEWCGROUP, CLONE_NEWUTS, CLONE_NEWIPC,
		CLONE_NEWUSER, CLONE_NEWPID, CLONE_NEWNET,
	} {
		fmt.Println(t.String())
	}
	// Output:
	// CLONE_NEWNS
	// CLONE_NEWCGROUP
	// CLONE_NEWUTS
	// CLONE_NEWIPC
	// CLONE_NEWUSER
	// CLONE_NEWPID
	// CLONE_NEWNET
}

func ExampleNamespaceType_Name() {
	fmt.Println(CLONE_NEWCGROUP.Name())
	// Output: cgroup
}

func ExampleNameToType() {
	for _, name := range []string{
		"mnt", "cgroup", "uts", "ipc", "user", "pid", "net", "spam",
	} {
		fmt.Printf("0x%08x\n", uint64(NameToType(name)))
	}
	// Output:
	// 0x00020000
	// 0x02000000
	// 0x04000000
	// 0x08000000
	// 0x10000000
	// 0x20000000
	// 0x40000000
	// 0x00000000
}

func ExampleIDwithType() {
	id, t := IDwithType("mnt:[12345678]")
	fmt.Printf("%q %d\n", t.Name(), id)
	// "nonsense" namespace textual representations return an identifier of
	// NoneID and a type of NaNS (not a namespace).
	id, t = IDwithType("foo:[-1]")
	fmt.Printf("%v %v\n", t, id)
	// Output:
	// "mnt" 12345678
	// NaNS NoneID
}
