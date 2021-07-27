/*

Package mounteneer allows accessing the file system contents from (other) mount
namespaces via procfs. If necessary, supporting "sandbox" processes are
temporarly deployed so that the current process can make good use of procfs root
"wormholes". This package normally offers less overhead compared to the original
gons/reexec method, as no marshalling and unmarshalling of information between a
parent and the re-executed child process is necessary anymore. Additionally,
when using a dedicated support "sandbox" binary, start time and resource
consumption is greatly reduced, too.

Mounteneers

Mounteneers kind of mount mount namespaces.

They abstract away the ugly details of when and how to make a mount namespace
(the "target mount namespaced") directly accessible from the current mount
namespace, via the proc file system. They even handle the totally bonkers
situation where a target mount namespace can only be referenced via a series of
bind-mounted mount namespace references (without any process using them at this
time).

Always remember: there are Engineers, Hellseneers, Mounteneers, ...

Do Not Leak

Make sure to always Close() any target mount namespace you (implicitly) opened
by creating a new Mounteneer. Failure to do so causes some mount namespaces to
become blocked against garbage collection by the Linux kernel until the
discovery process finally exits (which might be "never" in case of a discovery
service).

Note: for a much more detailed technical background please see the later
technical details at the end of this module documentation.

Use Cases

Let's look at the various use cases...

Note: For this and the following discussion we need to differentiate between:

  * the (path) reference to a mount namespace,
  * the paths to mount namespace contents.

The content paths are translated paths that allow a process (and OS-level
thread/goroutine) outside the target mount namespace to access the contents
inside(!) the target mount namespace. To drive home this crucial point, we
differentiate between referencing(!) the mount namespace and accessing(!) the
contents of that mount namespace.

Target Mount Namespace with Process Already Attached

Let's start with the simplest case, where we can not only reference the target
mount namespace directly, yet also directly access the target's mount namespace
contents.

This is the case when there's a process attached to the target mount namespace.
Here, we assume that the proc file system has been mounted in the context of the
initial PID namespace. This allows a discovery process, given proper capabilites
and CAP_SYS_PTRACE in particular, to access the target mount namespace. For
instance, we want to access mount namespace contents given the following
reference:

  * reference: /proc/12345/ns/mnt

Whenever a goroutine wants to access the contents in the target mount namespace
where process 12345 is attached to, it "just" needs to address it via the
indirection of "/proc/12345/root/" (for technical background details please see
later).

  * contents: /some/path/name → /proc/12345/root/some/path/name

Of course, we need to evaluate symbolic links in order to properly handle them
inside the "sandbox" of /proc/12345/root. We rely on
https://github.com/thediveo/procfsroot for the dirty details when it comes to
shambolic links.

Target Mount Namespace with Bind-Mounted Reference and No Process

Now we need to deal with target mount namespace references in form of
bind-mounts and where currently no process is attached to the target mount
namespace. Without the bind-mount such a namespace would quickly be garbage
collected by the Linux kernel.

  * reference: /bind-mounted/namespace/path

Admittedly, the situation can look slightly confusing. The above path is the
reference to the target mount namespace itself. But it cannot be used to address
the file system contents(!) inside that target mount namespace.

So, in order to allow a goroutine to access the contents in such a process-less,
bind-mounted mount namespace, we need to create a "pause" process just to get an
incredibly useful access path in the form of...

  * contents: /some/path/name → /proc/[(PAUSE)PID]/root/some/path/name

...where PAUSEPID is the PID of an auxiliary pause process we need to create for
the time a goroutine needs to access the contents inside the target mount
namespace.

Target Mount Namespace Referenced by Series of Bind-Mounts

Okay, this is now slightly going overboard. But didn't we claim "In every nook
and cranny"? Oh, yes, we did. So how do we deal with a bind-mounted mount
namespace reference that comes from another bind-mounted mount namespace and we
want to access the contents in the file system of the (final) target mount
namespace?

  * references: (1) /bind-mounted/ns1, (2) /bind-mounted/ns2

We interpret the (1) first reference in the context of the initial mount
namespace. Each following reference then is interpreted in the context of the
previously referenced mount namespace. Thus, (2) /bind-mounted/ns2 is a
reference in the file system ("contents") of the mount namespace referenced by
/bind-mounted/ns1.

In consequence, we start the same way as before when facing a bind-mounted
reference: we need a pause process in order to be able to access the next
bind-mounted reference. Thus, the next bind-mount reference in sequence is the
contents of the previous mount namespace. The first reference is always
interpreted in the context of the initial mount namespace; this matches with how
the discovery process for bind-mounted namespaces work.

  1. reference: (1) /bind-mounted/ns1
  2. "bindmount contents": /bind-mounted/ns2 → /proc/1/root/bind-mounted/ns2
  3. reference: (2) /bind-mounted/ns2
  4. "bindmount contents": /bind-mounted/ns2 →
     proc/[PAUSEPID1]/root/bind-mounted/ns2
  5. contents: /some/path/name → /proc/[PAUSEPID2]/root/some/path/name

Respectively, PAUSEPID1 and PAUSEPID2 are the PIDs of required pause processes
necessary in order to access the contents of process-less bind-mounted mount
namespaces.

Wormholes

The key to the inner workings of the mounteneer package is to know that Linux
allows to access the files and directories inside a different mount namespace
via the proc file system. This requires to be in a parent or same PID namespace
as the processes of which we want to access their mount namespaces. Typically,
this will be the initial PID namespace with a full view on each and every
process in the system.

And no, the "/proc/[PID]/root" paths are neither 0-days nor oversights; they
predate mount namespaces by some Linux eons and originate in those far and
light-insuffient ages of chroot(). And there's clear documentation of this
behavior in man 5 proc. Access to "/proc/[PID]/root" requires the OS-level
thread to posses the effective CAP_SYS_PTRACE capability.

Sandbox Process

The aforementioned "sandbox" processes are dummy processes that are doing
nothing except sleeping and hopefully using as few system resources as possible.
They vaguely resemble Kubernetes sandboxes, that's there they get their name
from.

The sole purpose of a sandbox process is to attach to a specified mount
namespace at start and then go to sleep, until killed. They're not supposed to
eat, and not RAM in particular.

With a sandbox process attached to a mount namespace it is now easy to access
the files and directories inside that mount namespace via the proc file system
and the "root" elements of processes in particular.

Compared to the "older" known full re-exec method (as implemented, for instance,
in thediveo/gons) in this package we don't have to access interesting
information from a re-executed child in a different mount namespace and then
painfully marshal this information back. Instead, when the sandbox correctly has
spun up, then we read everything from within our current threads/goroutines.

*/
package mounteneer
