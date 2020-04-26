/*

Package relations provides a Golang-idiomatic API to Linux query operations on
namespaces (see also: http://man7.org/linux/man-pages/man2/ioctl_ns.2.html). For
instance, for querying the parent namespace of a PID or user namespace, getting
the ID of a namespace, et cetera.

Linux-kernel namespaces can be referenced by filesystem path, open file
descriptor, or *os.File. This package therefore defines the following three
namespace reference types:

    * NamespacePath
    * NamespaceFd
    * NamespaceFile

All three types of namespace references define the following query operations:

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

*/
package relations
