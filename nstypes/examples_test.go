package nstypes

import (
	"fmt"
)

func ExampleTypeName() {
	fmt.Printf(TypeName(CLONE_NEWNS))
	// Output: mnt
}

func ExampleNamespaceType_String() {
	fmt.Println(CLONE_NEWUSER.String())
	// ...which can be simplified, because Println tries to String()ify its
	// arguments:
	fmt.Println(CLONE_NEWCGROUP)
	// Output: user
	// cgroup
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
