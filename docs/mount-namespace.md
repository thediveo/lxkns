# Mount Namespaces View

## Mount Point Hierarchy

For mount namespaces, **lxkns** not only shows the discovered mount namespaces
but additionally the mount points inside them.

![mount view](_images/mntview.png ':class=framedscreenshot')

> [!NOTE] **lxkns** discovers mount points even for process-less bind-mounted
> mount namespaces (...now that's quite a mouth full of techno babble). It does
> so by creating a temporary process attached to the bind-mounted mount
> namespace … et voilà: it sees the mount points inside that mount namespace.

Anyone who ever worked with `/proc/self/mountinfo` will probably appreciate this
neatly organized view: all discovered mount points are properly arranged into a
tree according to their **mount point paths** (VFS paths), starting from the
root ❶.

> [!WARNING] Do not confuse the mount point hierarchy with the mount point path
> hierarchy: while the latter bases on the VFS paths, the mount point hierarchy
> reflects the dependency of of mount points on each other and determining
> visibility and other properties.

A number in square brackets ❷ indicates the total amount of child and
grand-child mount points below a particular mount point.

A tree node with a folder icon and its name in italics ❸ is not a mount point,
but instead a common mount path element of several mount point paths in the tree view.

![mount point information](_images/mntinfo.png ':class=framedscreenshot')

- ❶ indicates a read-only mount point.
- ❷ is the type file system mounted at this point.
- ❸ signals the [propagation
  type](https://man7.org/linux/man-pages/man7/mount_namespaces.7.html#SHARED_SUBTREES):
  in this example, the mount point shares events with peer mounts as well as
  with slave mounts.
- ❹ opens a dialog showing further details for this mount point.

## Mount Point Details

Clicking or tapping the three dots "⋯" on a mount point node opens a pop up with
detail information about this mount point.

![mount point details](_images/mountpoint-details.png ':class=framedscreenshot')

- ❶ signals that this mount point is read-only (controlled by the mount
  options).
- ❷ is the VFS mount path...
- ...as opposed to the root path ❸ inside the file system getting mounted.
- ❹ shows the mount point propagation type as well as the peers, masters, and/or
  slaves to which mount events get propagated to. lxkns automaticaly translates
  the IDs into more descriptive mount namespace information and VFS mount paths.

## Overmounts/Hidden Mount Points

Mount points can become hidden in the VFS, either because a mount point later
gets "overmounted" at the same VFS path or because of a later mount point higher
up the VFS path hierarchy. For example:

![hidden mount point](_images/mnthidden.png ':class=framedscreenshot')

- mount point ❶ is hidden due to a later "overmount" ❷ with the same VFS path.
- mount point ❷ hides the previous mount point in the same VFS place (path).
