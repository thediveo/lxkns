/*

Package mounts enhances the Linux kernel's mountinfo data model
("/proc/[PID]/mountinfo") with mount point visibility ("overmounts") and a
hierarchical mount path tree.

Lack of visibility for a specific mount point indicates that it has been
rendered hidden by either another mount higher up the mount path hierarchy, such
as when "/a" hides "/a/b". Or a mount point is hidden by an "in-place" (bind)
mount point, such as when mounting "/a" on itself (for instance, to remove child
mounts or to change mount options).

For understanding this package it is important to understand the difference
between "mount paths" as opposed to "mount points".

Mount Paths

A "mount path" is the path of a mount point in the Virtual File System (VFS) of
Linux, such as "/a/b/c". The important take-away here is that a mount path is
not the same as mount point. A mount point has a mount path, but a mount path
isn't a mount point. In particular, there can be multiple mount points with the
same mount path and our our mount path data model thus dul(l)y reflects this.

Mount Points

In our context, a "mount point" describes the properties of a particular mount,
such as where the "files & directories" come from, mount options, and at which
path the mount point appears in the Virtual File System.

Mount points have uniques (integer) identifiers; however, while these are
unambiguous at a given time, they can be reused by the Linux kernel, so they're
not necessarily unambiguous over time.

Hidden Mounts and Overmounts

The terms "hidden mounts" and "overmounts" are used (more or less synonymously)
to describe mount points that have become inacessible from the perspective of
the VFS mount paths. For instance, because due to a new mount point higher up
the mount path hierarchy or by overmounting a mount point at the same mount path
with itself.

References

procfs(7) and there details about /proc/[PID]/mountinfo in particular:
https://man7.org/linux/man-pages/man5/procfs.5.html

*/
package mounts
