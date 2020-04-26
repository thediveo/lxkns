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

NamespaceFile embeds os.File (note that it does not embed a pointer, but os.File
directly). To create a *NamespaceFile from an *os.File, such as returned by
os.Open(), simply use the NewNamespaceFile() wrapper:

    nsf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))

The rationale here is to model NamespaceFile as close as possible to os.File,
and this implies that it should not be possible to create a NamespaceFile from a
nil *os.File.

Please note that NewNamespaceFile() expects two parameters, an *os.File as well
as an error. Simply specify a nil error in code contexts where there is clear
that the *os.File is valid and there was no error in getting it.

*/
package relations
