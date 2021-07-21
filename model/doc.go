/*

Package model defines the core of lxkns information model: Linux kernel
namespaces and processes, and how they relate to each other; with the additional
missing link between processes and user-land containers.

Linux namespaces partition certain OS resources and thus come in different
types. At the moment, there are namespaces for partitioning cgroups, IPC,
mounts, networks, PIDs, (monotonic) time, users, and UTS-related information
(hostname, ...).

Namespaces have unique identifiers, but these are not names, but inode numbers
(ignoring here the lost cause of device numbers on purpose).

Two types of namespaces are hierarchical: PID and user namespaces; all other
types of namespaces are "flat" without any hierarchy defined within namespaces
of the same type.

All namespaces are additionally owned by one user namespace or another. In case
of user namespaces this ownership actually is the parent-child namespace
relationship instead.

Namespaces may exist with processes, but also without any processes. The latter
requires references to such a namespace in form of either bind mounts or
parent-children relationships.

The lxkns information model shows which processes are currently "attached" to a
specific namespace, if any. However, to reduce noise, the information model only
references the "top-most" processes attached to a namespace, and leadership is
simply based on the process tree. These "top-most" processes are also dubbed
"leaders", and there's even a most senior leader process, the "ealdorman", based
on its starting time since the Boot Epoch.

All other processes also attached to a specific namespace can then be found by
following the process parent-child relationships, starting from the leader
processes.

The lxkns information model thus also contains the parent-child relationships
between processes. In addition, lxkns also models how individual processes are
attached to namespaces, so it's easy to quickly navigate forth and back between
namespaces and processes. Each process is always attached to exactly one
namespace of each type. However, "older" kernels lack time namespace support, so
be prepared that references to time namespaces will be nil on these kernels.

Moreover, leader processes are related to (userland) containers, where
applicable. Containers are also organized according to their managing container
engine. Of course, depending on host configuration, multiple container engines
might be present at the same time. A typical example is a Docker engine (daemon)
together with a containerd engine.

*/
package model
