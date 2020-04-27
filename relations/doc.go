/*

Package relations provides a Golang-idiomatic API to Linux query operations on
namespaces (see also: http://man7.org/linux/man-pages/man2/ioctl_ns.2.html). For
instance, for querying the parent namespace of a PID or user namespace, getting
the ID of a namespace, et cetera.

A particular Linux-kernel namespace can be referenced by a filesystem path, an
open file descriptor, or an *os.File. Thus, this package defines the following
three namespace reference types:

    * NamespacePath
    * NamespaceFd
    * NamespaceFile

All three types of namespace references define the following query operations
from the Relation interface:

    * ID() returns the ID of the referenced namespace.
    * User() returns the user namespace owning the referenced namespace.
    * Parent() returns the parent namespace of the referenced namespace.
    * OwnerUID() returns the UID of the owner of the referenced namespace.
    * Type() returns the type of referenced namespace; CLONE_NEWNS, ...

NamespacePath and NamespaceFd can be easily converted from or to string and
uintptr respectively.

    netns := NamespacePath("/proc/self/ns/net")
    path := string(netns)

As NamespaceFile mirrors os.File it cannot be directly converted in the way
NamespacePath and NamespaceFd can. However, things are not overly complex either
when keeping the following conversion examples in mind. To create a
*NamespaceFile from an *os.File, such as returned by os.Open(), simply use the
NewNamespaceFile() wrapper:

    nsf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))

The rationale here is to model NamespaceFile as close as possible to os.File,
and this implies that it should not be possible to create a NamespaceFile from a
nil *os.File.

Please note that NewNamespaceFile() expects two parameters, an *os.File as well
as an error. Simply specify a nil error in code contexts where there is clear
that the *os.File is valid and there was no error in getting it.

Getting back an *os.File in case it is explicitly required is also simple:

    f := &nsf.File

There's no need to panic because NamespaceFile embeds os.File, as opposed to
*os.File, on purpose: os.File is a struct which consists solely of a single
*os.file pointer to an implementation-internal structure. By embedding the outer
struct instead of a pointer to it we mimic the original handling as close as
possible, avoiding situations where a non-nil NamespaceFile points to a nil
*os.File.

Namespace IDs Without Device Numbers

Our package on purpose only returns the inode number of a namespace as its ID,
without the filesystem device number.

The reason lies in the way the Linux kernel manages namespaces: the IDs/inode
numbers of namespaces are solely managed through the so-called "nsfs"
filesystem. This "nsfs" filesystem is special in that it does not get listed as
an available filesystem in /proc/filesystems, and (in consequence) it cannot be
mounted at all. Instead, the kernel mounts it during startup automatically.
There is always only exactly one instance of nsfs, period. Thus, the device
number will be constant during the lifetime of a running Linux kernel and there
is no need to store and shuffle around an otherwise completely useless device
number, complicating the ID handling considerably ... as other namespace-related
packages show for worse.

With the knowledge about how the nsfs works under our belts comparing namespaces
for equality is as simple as:

    ns1.ID() == ns2.ID()

No special Equal() methods, nothing, nada, zilch. Just plain "==".
*/
package relations
