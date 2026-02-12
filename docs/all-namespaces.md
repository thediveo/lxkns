# All Namespaces View

The default view is named "all namespaces" and shows all discovered namespaces,
organized along the hierarchy of user namespaces ➊. The number next to the title
"Linux Namespaces" indicates to number of namespaces shown.

![view all namespaces](_images/allview.png ':class=framedscreenshot')

This view reflects the architectural design of the Linux namespaces, where any
namespace is always owned by a user namespace. In case of user namespaces owning
user namespaces this is also the parent-child relationship.

lxkns shows for each user namespace ➊ the "most senior" process with its name
and PID ➋, as well as the user ID and user name to which this user namespace
belongs. The "most senior" process is also termed the "ealdorman"; it is the
topmost *and oldest* process in the process tree that is still attached to the
user namespace.

Now, there might be groups of processes attached to the user namespace with
different cgroup controllers ➌, ➍. These are additionally listed as to not miss
such different "tenants".

> [!NOTE] The init process with PID 1 is always shown first, while all other
> "tenants" are sorted by their process names. Additionally, PID 1 is marked
> with a golden crown icon.

Those namespaces ➌ created at system start are called "initial namespaces".
These are visually marked by dashed borders to make them easily spottable.

Other tenants ➍ might either use some or all of the existing namespaces ("shared
namespaces") or newly created namespaces instead. In case a namespace is
"shared" it is shown washed out. "Reused" namespaces can be hidden in the
[settings](#settings).
