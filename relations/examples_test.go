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

func Example_iD() {
	id, _ := NamespacePath("/proc/self/ns/net").ID()
	fmt.Println("id of my network namespace:", id)
}

func Example_user() {
	userns, _ := NamespacePath("/proc/self/ns/net").User()
	id, _ := userns.ID()
	userns.Close()
	fmt.Println("user namespace id owning my network namespace:", id)
}

func Example_ownerUID() {
	uid, _ := NamespacePath("/proc/self/ns/user").OwnerUID()
	fmt.Println("user namespace id owning my network namespace:", uid)
}

func Example_parent() {
	parentuserns, _ := NamespacePath("/proc/self/ns/user").Parent()
	id, _ := parentuserns.ID()
	parentuserns.Close()
	fmt.Println("parent user namespace id of my user namespace:", id)
}

func Example_type() {
	nstype, _ := NamespacePath("/proc/self/ns/pid").Type()
	fmt.Printf("0x%08x\n", uint(nstype))
	fmt.Println(nstype.String())
	// Output:
	// 0x20000000
	// CLONE_NEWPID
}
