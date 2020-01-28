package relations

import (
	"fmt"

	"github.com/thediveo/lxkns/nstypes"
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
	fmt.Println(nstypes.TypeName(nstype))
	// Output:
	// 0x20000000
	// pid
}

func ExampleTypeName() {
	fmt.Println(nstypes.TypeName(nstypes.CLONE_NEWNET))
	// Output: net
}
