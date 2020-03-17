/*

pidtree displays a tree of processes within their PID namespaces, as well as
the local PIDs of processes (where applicable).

The pidtree command bears some resemblance with the pstree command
(http://man7.org/linux/man-pages/man1/pstree.1.html) in that both display a
tree of processes. However, pidtree focuses on how processes are organized in
PID namespaces; something pstree isn't aware of.

Usage:

  pidtree [flag]

The flags are:

  ...

Display:

The process tree starts at the topmost PID namespace; when started in the
initial PID namespace, then this will be the initial PID namespace.

  pid:[4026531836], owned by UID 0 ("root")
  ├─ "systemd" (1748)
  [...]

When started in a child PID namespace, then the topmost PID namespace will be
that child PID namespace. The Linux kernel makes it impossible for processes
to reach out into parent (or sibling) PID namespaces, thus pidtree cannot show
proper PID namespacing information for such processes above the starting
point's PID namespace (please also see below).

Whenever a child process lives in a different PID namespace than its parent
process, pstree shows an intermediate PID namespace node between parent and
child process(es). These PID namespace nodes show the namespace ID (inode
number), as well as the user ID and user name "owning" the PID namespace. For
a PID namespace, the owner is the user which created the user namespace, which
in turn was active when the PID namespace was created. Or to phrase this chain
slightly differently: the PID namespace is owned by a user namespace, and that
user namespace is owned by a user.

  pid:[4026531836], owned by UID 0 ("root")
  ├─ "systemd" (1)
  │  ├─ "systemd-journal" (417)
  [...]
  │  │  └─ "unshare" (5309)
  │  │     └─ pid:[4026532229], owned by UID 1000 ("thediveo")
  │  │        └─ "bash" (5310=1)
  │  │           └─ "unshare" (5344=24)
  │  │              └─ pid:[4026532247], owned by UID 1000 ("thediveo")
  │  │                 └─ "bash" (5345=1)
  │  │                    └─ "sleep" (5529=25)
  [...]

For PID namespaces different from the starting PID namespace, pidtree not only
shows the process PIDs as seen from the starting PID namespace. Additionally,
it also shows the PID from the perspective of the "local" PID namespace. The
"local" PID namespace is the PID namespace a process is joined to. The "local"
PID is thus the PID seen by the process itself. For instance, this information
about the local PID might be useful when entering child PID namespaces and
controlling childs or diagnosing logs with PIDs.

  │  │              └─ pid:[4026532247], owned by UID 1000 ("thediveo")
  │  │                 └─ "bash" (5345=1)
  │  │                    └─ "sleep" (5529=25)

Insufficient Privileges/Capabilities:

When pidtree is started without the necessary privileges (in particular, the
CAP_SYS_PTRACE capability) it has only limited visibilty onto Linux kernel
namespaces. In this situation, pidtree will only show processes which are
either in the same PID namespace as itself, or which are direct or indirect
children of such processes. When a child process is in an unknown PID
namespace, then it will be prefixed with a "pid:[???]" indication and
additionally an unknown local PID "=???" will be shown.

  [...]
  │  ├─ "bash" (14725)
  │  │  └─ pid:[???] "sudo" (14738=???)
  │  │     └─ pid:[???] "unshare" (14742=???)
  │  │        └─ pid:[???] "bash" (14744=???)
  │  │           └─ pid:[???] "unshare" (14756=???)
  │  │              └─ pid:[???] "bash" (14757=???)
  │  │                 ├─ pid:[???] "sleep" (14773=???)
  │  │                 └─ pid:[???] "bash" (15662=???)
  [...]

*/
package main
