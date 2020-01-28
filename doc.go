/*

Package lxkns discovers Linux kernel namespaces, such as user, PID, network,
and the other types of namespaces. This package can discover namespaces not
only when processes have joined them, but also when they have been
bind-mounted or are still referenced by process file descriptors. Also, for
PID and user namespaces, their hierarchies are discovered (unless running on
an ancient kernel). Moreover, for user namespaces their owning user and the
owned namespaces will be discovered too.

And finally, namespaces can be related to leading (or "root") processes joined
to them, based on the process tree.

The discovery process can be controlled in several aspects, according to the
range of discovery of namespace types and places to search namespaces for,
according to the needs of users of the lxkns package.

*/
package lxkns
