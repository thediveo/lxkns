/*

Package ops provides a Golang-idiomatic API to query and switching operations on
Linux-kernel namespaces.

Namespace Queries

A particular Linux-kernel namespace can be referenced by a filesystem path, an
open file descriptor, or an *os.File. Thus, this package defines the following
three namespace reference types:

    * NamespacePath
    * NamespaceFd
    * NamespaceFile

All three types of namespace references define the following query operations
from the Relation interface (which map to a set of ioctl() calls, see:
http://man7.org/linux/man-pages/man2/ioctl_ns.2.html, with the exception of the
ID query):

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

Switching Namespaces

Switching namespaces is a slightly messy business in Golang: it is subject to
both Golang runtime limitations as well as Linux kernel restrictions imposed
especially on multi-threaded processes. In particular, after the Golang runtime
has started, threads cannot change their user namespaces and mount namespaces
anymore. Also, processes in general cannot switch themselves into a different
PID namespace, but only their future child processes. Luckily, switching other
types of namespaces is less restricted, such as switching a specific Go routine
(rather, its locked OS thread) into another network namespace (and back again)
is almost painless. However, OS threads need to hold both sufficient effective
privileges for themselves as well as they must have sufficient (evaluated)
capabilities in the namespace to switch to, please see
http://man7.org/linux/man-pages/man2/setns.2.html and
http://man7.org/linux/man-pages/man7/user_namespaces.7.html for details about
the specific capabilities needed and how capabilities of a process with relation
to a destination namespace are evaluated.

The Go() function runs a function as a Go routine in the specified namespace(s).
It returns an error in case switching into the specified namespaces fails,
otherwise it simply returns nil. Please note that Go() doesn't call the
specified function synchronously, but instead as a new Go routine.

    netns := ops.NamespacePath("/proc/self/ns/net")
    if err := ops.Go(func() {
        fmt.Println("Nobody expects the Spanish Inquisition!")
    }, netns)

While this might seem inconvenient at first, this design actually is very robust
in view of any problems that might pop up when trying to switch the current
(locaked) OS thread back into its original namespaces; the OS thread would be
unrecoverable, but without a way to disassociate its Go routine from it. The
Go() function avoids this situation by executing the desired function in a
throw-away Go routine, so the Golang runtime can easily throw away the tainted
OS thread which is locked to it afterwards. Sometimes, throwing things away is
much cleaner (and not only for certain types of PPE).

If a Golang process needs to switch mount, PID, and user namespaces, we
recommend using the gons package https://github.com/thediveo/gons in combination
with its reexec subpackage (gons provide namespace switching before the Golang
runtime starts, while reexec forks a Golang process and reexecutes it, with the
reexecuted child then runnining a specific function only in the specified
namespaces).

Namespace IDs

This package works with namespace identifiers in the form of tuples made from
the inode number of a namespace and the associated filesystem device number
(device ID). While at this time the inode number would be sufficient, as
currently all namespaces are solely managed through the so-called "nsfs"
filesystem, the tuple-based model follows the advice from
http://man7.org/linux/man-pages/man7/namespaces.7.html as well as the dire
kernel developer warning to create more havoc by deploying multiple namespace
filesystem instances.

Comparing namespaces for equality is as simple as, as long both identifiers come
from an origin honoring the device IDs:

    if ns1.ID() == ns2.ID() {}

No special Equal() methods, nothing, nada, zilch. Just plain "==".

Unfortunately, the same kernel devs ignored their own warnings and happily
output any namespace textual reference using only the inode number, such as in
`net:[4026531905]`. And they left it to us to face the music they're playing;
CLI tools so far only use the kernel's incomplete textual format. In
consequence, species.IDwithType("net:[...]") returns only incomplete namespace
identifier information.

To compare two namespace identifiers, where one might be incomplete:

    if ns1.ID().SloppyEqual(ns2.ID()) {}

If both namespace IDs have non-zero device IDs, then SloppyEqual works the same
as "==", doing a full check for equality.

To look up a namespace by incomplete ID, use:

    allns.Namespaces[lxkns.PIDNS].SloppyByIno(ns1.ID())

Please be aware that this method currently works by assuming that there
currently is only a single `nsfs` instance and then taking the missing device ID
from an arbitrary namespace map entry. However, if the dire warning might come
true in the future, then the implementation of SloppyByIno() will be upgraded
accordingly (together with IDwithType).

*/
package ops
