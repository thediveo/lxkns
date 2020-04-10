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

package relations

import (
	"fmt"
)

func ExampleID() {
	id, _ := ID("/proc/self/ns/net")
	fmt.Println("id of my network namespace:", id)
}

func ExampleUser() {
	userns, _ := User("/proc/self/ns/net")
	id, _ := ID(userns)
	userns.Close()
	fmt.Println("user namespace id owning my network namespace:", id)
}

func ExampleOwnerUID() {
	uid, _ := OwnerUID("/proc/self/ns/user")
	fmt.Println("user namespace id owning my network namespace:", uid)
}

func ExampleParent() {
	parentuserns, _ := Parent("/proc/self/ns/user")
	id, _ := ID(parentuserns)
	parentuserns.Close()
	fmt.Println("parent user namespace id of my user namespace:", id)
}

func ExampleType() {
	nstype, _ := Type("/proc/self/ns/pid")
	fmt.Printf("0x%08x\n", uint(nstype))
	fmt.Println(nstype.String())
	// Output:
	// 0x20000000
	// CLONE_NEWPID
}
