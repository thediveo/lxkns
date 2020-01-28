/*

Package relations provides a Golang-based API to Linux query operations on
namespaces (see also: http://man7.org/linux/man-pages/man2/ioctl_ns.2.html).
The following query operations are supported: User() getting the user
namespace owning a namespace, Parent() returning the parent of a pid or user
namespace, OwnerUID() revealing the creator's UID of a user namespace, and the
Type() of a namespace.

These API functions all take a namespace reference. They accept these
references in one out of three different forms: (1) as file path strings, (2)
as os.File, and finally (3) as file descriptor numbers (ints).

*/
package relations
