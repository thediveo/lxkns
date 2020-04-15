/*

Package lxkns discovers Linux kernel namespaces (of types cgroup, ipc, mount,
net, pid, user, and uts). This package discovers namespaces not only when
processes have joined them, but also when namespaces have been bind-mounted or
are only referenced anymore by process file descriptors.

In case of PID and user namespaces, lxkns additionally discovers their
hierarchies, except when running on a really ancient kernel before 4.9.
Furthermore, for user namespaces the owning user ID and the owned namespaces
will be discovered too.

And finally, lxkns relates namespaces to the "leading" (or "root") processes
joined to them; this relationship is basically derived from on the process tree
hierarchy.

The namespace discovery process can be controlled in several aspects, according
to the range of discovery of namespace types and places to search namespaces
for, according to the needs of API users of the lxkns package.

Discovery

Running a namespace discovery is a single function call (and an initial one-time
support function call).

    import (
        "github.com/thediveo/gons/reexec"
        "github.com/thediveo/lxkns"
    )

    func main() {
        reexec.CheckAction()
        ...
        allns := lxkns.Discover(lxkns.FullDiscovery)
        ...
    }

Technical note: in order to discover namespaces in some locations, such as
bind-mounted namespaces, lxkns needs to fork the process it used from in, in
order to switch the forked copy into other mount namespaces for further
discovery. In order to implement this mechanism as painless as possible, process
using lxkns need to call reexec.CheckAction() as early as possible from their
main().

Information Model, Base Level

Not totally unexpectedly, the lxkns discovery information model at its most
basic level comprises namespaces. In the previous code snippet, the information
model returned is stored in the "allns" variable for further processing. The
result organizes the namespaces found by type. For instance, the following code
snippet prints all namespaces, sorted first by type and then by namespace
identifier:

    // Iterate over all 7 types of Linux-kernel namespaces, then over all
    // namespaces of a given type...
    for nsidx := lxkns.MountNS; nsidx < lxkns.NamespaceTypesCount; nsidx++ {
        for _, ns := range allns.SortedNamespaces(nsidx) {
            println(ns.Type().Name(), ns.ID())
        }
    }

Because namespaces have no order defined, the discovery results "list" the
namespaces in per-type maps, indexed by namespace identifiers. For convenience,
SortedNamespaces() returns the namespaces of a specific type in a slice instead
of a map, sorted numerically by the namespace identifiers.

Technically, these namespace identifiers are 64bit unsigned inode numbers and
come from the special "nsfs" namespace filesystem integrated with the Linux
kernel. And before someone tries: nope, the nsfs cannot be mounted; and it even
does not appear in the kernel's list of namespaces.

Unprivileged Discovery and How To Not Panic

While it is possible to discover namespaces without root privileges, this won't
return the full set of namespaces in a Linux host. The reason is that while an
unprivileged discovery is allowed to see some basic information about all
processes in the system, it is not allowed to query the namespaces such
privileged processes are joined too. In addition, an unprivileged discovery may
turn up namespaces (for instance, when bind-mounted) for which the identifier is
discovered, but further information, such as the parent or child namespaces for
PID and user namespaces, is undiscoverable.

Users of the lxkns information model thus must be prepared to handle incomplete
information yielded by unprivileged lxkns.Discover() calls. In particular,
applications must be prepared to handle:

  * more than a single "initial" namespace per type of namespace,
  * PID and user namespaces without a parent namespace,
  * namespaces without owning user namespaces,
  * processes not related to any namespace.

In consequence, always check interface values and pointers for nil values like a
pro. You can find many examples in the sources for the "lsuns", "lspidns", and
"pidtree" CLI tools (inside the cmd sub-package).

In-Capabilities

It is possible to run full discoveries without being root, when executing the
discovery process with the following effective capabilities:

  * CAP_SYS_PTRACE -- no joking here, that's what needed for reading namespace refs
  * CAP_SYS_CHROOT -- for mount namespace switching
  * CAP_SYS_ADMIN  -- for mount namespace switching

Considering that especially CAP_SYS_PTRACE being essential there's probably not
much difference to "just be root" in the end, unless you want show off your
"capabilities capabilities".

Namespace Hierarchies

PID and user namespaces form separate and independent namespaces hierarchies.
This parent-child hierarchy is exposed through the lxkns.Hierarchy interface of
the discovered namespaces.

Please note that lxkns represents namespaces often using the lxkns.Namespace
interface when the specific type of namespace doesn't matter. In case of PID and
user-type namespaces an lxkns.Namespace can be "converted" into an interface
value of type lxkns.Hierarchy using a type assertion, in order to access the
particular namespace hierarchy.

    // If it's a PID or user namespace, then we can turn a "Namespace"
    // into an "Hierarchy" in order to access hierarchy information.
    if hns, ok := ns.(lxkns.Hierarchy); ok {
        if hns.Parent() != nil {
            ...
        }
        for _, childns := range hns.Children() {
            ...
        }
    }

Ownership

User namespaces play the central role in controlling the access of processes to
other namespaces as well as the capabilities process gain when allowed to join
user namespaces. A comprehensive discussion of the rules and their ramifications
is beyond this package documentation. For starters, please refer to the man page
for user_namespaces(7), http://man7.org/linux/man-pages/man7/user_namespaces.7.html.

The controlling role of user namespaces show up in the discovery information
model as owner-owneds relationships: user namespaces own non-user namespaces.
And non-user namespaces are owned by user namespaces, the "ownings". In case you
are scratching your head why the Gopher the owned namespaces are related to as
"ownings": welcome to the wonderful Gopher world of "er"-ers, where interface
method naming conventions create wonderful identifier art.

If a namespace interface value represents a user-type namespace, then it can be
"converted" into an lxkns.Ownership interface value using a type assertion. This
interface discloses which namespaces are owned by a particular user namespace.
Please note that this includes child user namespaces, too.

    // Get the user namespace -owned-> namespaces relationships.
    if owns, ok := ns.(lxkns.Ownership); ok {
        for _, ownedns := range owns.Ownings() {
            ...
        }
    }

In the opposite direction, the owner of a namespace can be directly queried via
the lxkns.Namespace interface:

    // Get the namespace -owned by-> user namespace relationship.
    ownerns := ns.Owner()

When asking a user namespace for its owner, the parent user namespace is
returned in accordance with the Linux ioctl()s for discovering the ownership of
namespaces.

Namespaces and Processes

The lxkns discovery information model also relates processes to namespaces, and
vice versa. After all, processes are probably the main source for discovering
namespaces.

For this reason, the discovery results (in "allns" in case of the above
discovery code example) not only list the namespaces found, but also a snapshot
of the process tree at discovery time (please relax now, as this is a snapshot
of the "tree", not of all the processes themselves).

    // Get the init(1) process representation.
    initprocess := allns.Processes[lxkns.PIDType(1)]
    for _, childprocess := range initprocess.Children() {
        ...
    }

Please note that the process tree information is for convenience; it's not a
replacement for the famous gopsutil package in many use cases. However, the
process tree information show which namespaces are used by (or "joined by")
which particular processes.

    // Show all namespaces joined by a specific process, such as init(1).
    for nsidx := lxkns.MountNS; nsidx < lxkns.NamespaceTypesCount; nsidx++ {
        println(initprocess.Namespaces[nsidx].String())
    }

It's also possible, given a specific namespace, to find the processes joined to
this namespace. However, the lxkns information model optimizes this relationship
information on the assumption that in many situations not the list of all
processes joined to a namespace is needed, but actually only the so-called
"leader" process or processes.

A leader process of namespace X is the process topmost in the process tree
hierarchy of processes joined to namespace X. It is perfectly valid for a
namespace to have more than one leader process joined to it. An example is a
container with its own processes joined to the container namespaces, and an
additional "visiting" process also joined to one or several namespaces of this
container. The lxkns information then is able to correctly handle and represent
such system states.

    // Show the leader processes joined to the initial user namespace.
    for _, leaders := range initprocess.Namespaces[lxkns.UserNS].Leaders() {
        ...
    }

*/
package lxkns
