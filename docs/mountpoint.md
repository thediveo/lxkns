# Mount Point Discovery

In mount namespaces, lxkns discovers the **mount point hierarchy** from
`mountinfo` in the [process file
system](https://man7.org/linux/man-pages/man5/proc.5.html). It then not only
derives the VFS mount **path hierarchy** from this information, but determines
also mount point **visibility**. This is a slightly involved process[^1], to put
it mildly, so lxkns does it for you.

## Point and Path Hierarchies

The mount point hierarchy is separate from the mount path hierarchy:

- the mount (point) **path hierarchy** is the tree of mount points along **VFS
  paths**. At any given VFS path name there might be none, one, or even several
  mount points (see "overmounts" below). This hierarchy answers the question:
  "which mount(s) are at a specific path?".

- the mount **point hierarchy** is the tree of mount points as managed by the
  Linux kernel and reflects both the history and hierarchy of mounts.

As the Linux kernel only returns the mount point hierarchy, lxkns runs a
slightly involved process in order to determine the correct visibility of mount
points.

## Overmounts and Visibility

Mount points can become hidden (invisible) when getting "overmounted":

- **in-place overmount**: another mount point at the same mount path as a
  previous mount point hides the former mount point. It is even possible to
  bind-mount a mount point onto itself, changing mount options, such as mount
  point propagation, et cetera.

- **overmount higher up the mount path**: a mount point has a prefix path of
  another mount path and mount point and thus is hidding the latter, including
  all mount points with paths further down the hierarchy below the hidden mount
  point.

Lxkns also discovers mount points in mount namespaces that currently are
process-less, but that have been bind-mounted into the VFS – one example is the
["snap" technology](https://snapcraft.io/docs) by Canonical.

#### Notes

[^1]: see also this StackExchange question: [Check for
      overmount?](https://unix.stackexchange.com/questions/398983/check-for-overmount);
      the originally proposed algorithm unfortunately causes false positives, so
      lxkns uses a completely redesigned algorithm instead.
