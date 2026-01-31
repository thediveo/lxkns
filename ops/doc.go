/*
Package ops provides a Golang-idiomatic API to the query and switching
operations on Linux-kernel namespaces, hiding ioctl()s and syscalls.

# Namespace Queries

A particular Linux-kernel namespace can be referenced by a filesystem path, an
open file descriptor, or an *[os.File]. Thus, this package defines the following
three (six) namespace reference types:

  - [NamespacePath] (and [TypedNamespacePath])
  - [NamespaceFd] (and [TypedNamespaceFd])
  - [NamespaceFile] (and [TypedNamespaceFile])

The only difference between the NamespaceXXX and TypedNamespaceXXX reference
types are: if the the type of namespace referenced is known beforehand, then
this knowledge might be used to either optimize [relations.Type] lookups, or
support Linux kernels before 4.11 which lack the ability to query the type of
namespace via an ioctl(). In particular, this allows using the [Visit] function
(see below) to be used on such older kernels.

All these types of namespace references define the following query operations
from the [relations.Relation] interface (which map to a set of ioctl() calls,
see [ioctl_ns(2)], with the exception of the ID query):

  - [relations.ID] returns the ID of the referenced namespace.
  - [relations.User] returns the user namespace owning the referenced namespace.
  - [relations.Parent] returns the parent namespace of the referenced namespace.
  - [relations.OwnerUID] returns the UID of the owner of the referenced
    namespace.
  - [relations.Type] returns the type of referenced namespace; CLONE_NEWNS, ...

NamespacePath and NamespaceFd can be easily converted from or to string and
uintptr respectively.

	netns := NamespacePath("/proc/self/ns/net")
	path := string(netns)

In case you want to use the [Visit] function for switching namespaces and you
need to support Linux kernels before 4.11 (which lack a required ioctl) then you
can resort to [TypedNamespacePath] instead of [NamespacePath].

	netns := TypedNamespacePath("/proc/self/ns/net", species.CLONE_NEWNET)

As [NamespaceFile] mirrors [os.File] it cannot be directly converted in the way
[NamespacePath] and [NamespaceFd] can. However, things are not overly complex
either when keeping the following conversion examples in mind. To create a
*NamespaceFile from an *os.File, such as returned by [os.Open], simply use the
[NewNamespaceFile] wrapper:

	nsf, err := NewNamespaceFile(os.Open("/proc/self/ns/net"))

The rationale here is to model [NamespaceFile] as close as possible to
[os.File], and this implies that it should not be possible to create a
NamespaceFile from a nil *os.File.

Please note that [NewNamespaceFile] expects two parameters, an *[os.File] as
well as an [error]. Simply specify a nil error in code contexts where there is
clear that the *os.File is valid and there was no error in getting it.

Getting back an *[os.File] in case it is explicitly required is also simple:

	f := &nsf.File

There's no need to panic because [NamespaceFile] embeds [os.File], as opposed to
*os.File, on purpose: os.File is a struct which consists solely of a single
*os.file pointer to an implementation-internal structure. By embedding the outer
struct instead of a pointer to it we mimic the original handling as close as
possible, avoiding situations where a non-nil NamespaceFile points to a nil
*os.File.

# Switching Namespaces

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
capabilities in the namespace to switch to, please see [setns(2)] and
[user_namespaces(7)] for details about the specific capabilities needed and how
capabilities of a process with relation to a destination namespace are
evaluated.

This package provides three means to execute some Go code in an OS thread with
namespaces switched as specified:

  - [Go](f, namespaces...) – asynchronous f in the specified namespaces.
  - [Execute](f, namespaces...) – synchronous f in the specified namespaces with
    result.
  - [Visit](f, namespaces...) – synchronous f in the specified namespaces in
    same Go routine.

These namespace-switching methods differ as follows: [Go](f, namespaces...) acts
very similar to the “go” statement in that it runs the given function f as a new
go routine, but with its executing OS thread locked and switched into the
specified namespaces.

[Execute](f, namespaces...) is a synchronous version of Go() which waits for the
namespace-switched f() to complete and to return some result (in form of an
any). Execute then returns this result to the caller.

[Visit](f, namespaces...) is for those situations where the caller wants to
avoid creating a new go routine, but is prepared to throw away its current go
routine in case Visit fails switching out of the namespaces afterwards, so the
current OS thread and its go routine is toast.

# Go

The [Go] function runs a function as a Go routine in the specified namespace(s).
It returns an error in case switching into the specified namespaces fails,
otherwise it simply returns nil. Please note that Go doesn't call the
specified function synchronously, but instead as a new Go routine.

	netns := ops.NamespacePath("/proc/self/ns/net")
	err := ops.Go(func() {
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
recommend using the [gons] package in combination with its reexec subpackage
(gons provide namespace switching before the Golang runtime starts, while reexec
forks a Golang process and reexecutes it, with the reexecuted child then
runnining a specific function only in the specified namespaces).

# Execute

[Execute] is the synchronous twin of Go(): it waits for the namespace-switched
function f() to complete and to return an any. Execute then passes on
this result to its caller.

	netns := ops.NamespacePath("/proc/self/ns/net")
	result, err := ops.Execute(func() any {
	    return "Nobody expects the Spanish Inquisition!"
	}, netns)

# Visit

If unsure, use [Go] or [Execute] instead. Only use [Visit] if you understand
that it can get you in really hot water and you are prepared to accept any
consequences.

In case a go routine wants to hop into a namespace and then out of it again,
without the help of a new go-routine, then [Visit] helps with that. However, due
to Golang's runtime design, if getting back to the original namespaces before
the call to Visit fails, then any such go routine must be prepare to sacrifice
itself, because by then it has a locked OS thread on its back in an unknown
namespace attachment state, and further namespace hopping might end badly.

If the caller is in a throw-away go routine itself and needs to run some code
synchronously in other namespaces, then [Visit] gives some optimization over
[Go] and especially [Execute], as it avoids spinning up another go routine.

For more background information please see [M0 is Special].

# Namespace IDs

This package works with namespace identifiers in the form of tuples made from
the inode number of a namespace and the associated filesystem device number
(device ID). While at this time the inode number would be sufficient, as
currently all namespaces are solely managed through the so-called “nsfs”
filesystem, the tuple-based model follows the advice from [namespaces(7)] as
well as the dire kernel developer warning to create more havoc by deploying
multiple namespace filesystem instances.

Comparing namespaces for equality is as simple as, as long both identifiers come
from an origin honoring the device IDs:

	if ns1.ID() == ns2.ID() {
		// ...
	}

No special Equal methods, nothing, nada, zilch: just plain "==".

Unfortunately, the same kernel devs ignored their own dire warnings and happily
output any namespace textual reference using only the inode number, such as in
“net:[4026531905]”. And they left it to us to face the music they're playing;
CLI tools so far only use the kernel's incomplete textual format. To ease a
future transition, [species.IDwithType]("net:[...]") returns complete namespace
identifier information by supplying the missing device ID for the current nsfs
filesystem itself.

In consequence, user code should avoid creating any [species.NamespaceID] directly, but
instead through [species.IDwithType], such as:

	nsid1, _ := species.IDwithType("net:[4026531905]")

or, given the inode as a number, not text, using [species.NamespaceIDFromInode]:

	nsid1 := species.NamespaceIDFromInode(4026531905)

Sigh.

[M0 is Special]: https://thediveo.github.io/#/art/namspill
[ioctl_ns(2)]: http://man7.org/linux/man-pages/man2/ioctl_ns.2.html
[setns(2)]: http://man7.org/linux/man-pages/man2/setns.2.html
[user_namespaces(7)]: http://man7.org/linux/man-pages/man7/user_namespaces.7.html
[gons]: https://github.com/thediveo/gons
[namespaces(7)]: http://man7.org/linux/man-pages/man7/namespaces.7.html
*/
package ops
