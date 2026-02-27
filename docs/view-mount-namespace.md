# Mount Namespaces View

For mount namespaces, **lxkns** not only shows the discovered mount namespaces
but additionally the mount points inside them. This is especially useful to nose
around how containers and [snaps](https://snapcraft.io/) are set up behind their
file system scenes.

## Mount Point Hierarchy

If you ever had to deal with `/proc/self/mountinfo` and felt the pain of trying
to understand it, then you will surely like that **lxkns** neatly displays the
discovered mount points in a tree, according to their **mount point paths** (VFS
paths) and starting from the root ➊.

![mount view](_images/mntview.png ':class=framedscreenshot')

> [!NOTE] **lxkns** discovers mount points even for process-less bind-mounted
> mount namespaces (...now that's quite a mouth full of techno babble). It does
> so by creating a temporary task or process attached to the bind-mounted mount
> namespace … et voilà: it sees the mount points inside that mount namespace.
> More details can be found in our section on [Mountineers](mountineers).

> [!WARNING] Do not confuse the "mount point hierarchy" with the "mount point
> _path_ hierarchy": while the latter _path_ hierarchy bases on the VFS paths,
> the hierarchy of mount points reflects the dependency of these mount points on
> each other and thus determining visibility and other properties.

A number in square brackets ➋ indicates the total amount of child and
grand-child mount points below a particular mount point.

A tree node with a folder icon and its name in italics ➌ is not a mount point,
but instead a common mount path element of several mount point paths in the tree view.

Individual mount points show additional information:

![mount point information](_images/mntinfo.png ':class=framedscreenshot')

- ➊ indicates a read-only mount point.
- ➋ is the type file system mounted at this point.
- ➌ shows the [propagation
  type](https://man7.org/linux/man-pages/man7/mount_namespaces.7.html#SHARED_SUBTREES):
  in this example, the mount point shares events with peer mounts as well as
  with slave mounts.
- ➍ opens a dialog showing further details for this mount point.

## Mount Point Details

Clicking or tapping the three dots "⋯" on a mount point tree node opens a pop up
with detail information about this mount point.

![mount point details](_images/mountpoint-details.png ':class=framedscreenshot')

- ➊ announces this mount point to be read-only (controlled by the mount
  options).
- ➋ is the VFS mount path...
- ...as opposed to the root path ➌ inside the file system getting mounted.
- ➍ shows the mount point propagation type as well as the peers, masters, and/or
  slaves to which mount events get propagated to. lxkns automaticaly translates
  the IDs into more descriptive mount namespace information and VFS mount paths.

## Overmounts/Hidden Mount Points

Mount points can become hidden in the VFS, either because a mount point later
gets "overmounted" at the same VFS path or because of a later mount point higher
up the VFS path hierarchy. For example:

![hidden mount point](_images/mnthidden.png ':class=framedscreenshot')

- mount point ➊ is hidden due to a later "overmount" ➋ with the same VFS path.
- mount point ➋ hides the previous mount point in the same VFS place (path).
