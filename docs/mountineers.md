# Mountineers

There are Engineers, Hellsineers[^1], Mountineers...

Okay, all the joking aside, "mountineers" are used in lxkns to access the file
system inside those mount namespaces which are only bind-mounted and thus
currently have no process attached to them. So, mountineers mount mount
namespaces in order to access their contents using ordinary file operations
(albeit not in the technical sense of the `mount(2)` syscall).

In fact, mountineers are used to unify access to the file system contents inside
mount namespaces in general, regardless of bind-mounted or not. The mountineers
hide all the logic to decide whether there's a convenient process already in
place that gives us access via the process filesystem, or whether we need to
spin up our own dedicates, yet temporary "sandbox" process (see also:
[mntnssandbox](mntnssandbox)).

## Usage

Given some mount namespace reference, simply create a mountineer and then access
the filesystem inside the mount namespace using translated paths. When done,
close the mountineer in order to release any sandbox process that might have
been needed.

```go
mnteer, err := mountineer.New(
    model.NamespaceRef{"/proc/1/ns/mnt", "/run/snapd/ns/chromium.mnt"}, nil)
defer mntneer.Close()
etchostname, err := mnteer.ReadFile("/etc/hostname")
```

In case the mountineer does not provide its own convenience version of a file or
directory operation, paths inside a target mount namespace can be translated
into paths usable from the program being in a different mount namespace. This is
termed "resolving" and uses the `Resolve()` method.

```go
mnteer, err := mountineer.New(
    model.NamespaceRef{"/proc/1/ns/mnt", "/run/snapd/ns/chromium.mnt"}, nil)
defer mntneer.Close()
etchostpath, err := mnteer.Resolve("/etc/hostname")
etchostname, err := ioutil.ReadFile(etchostpath)
```

## Technical Background

Mountineers work around the limitation that multi-threaded processes (and thus
Go programs) cannot switch mount namespaces.

**Before mountineers**, lxkns forked its (parent) process and then re-executed
itself in order to switch into a target mount namespace before the go runtime
spins up and called a registered "action" function (leveraging the
[thediveo/gons](https://github.com/TheDiveO/gons) module). This design has the
important drawback that optional call parameters as well as any results need to
be channelled forth and back between the parent and re-executed child processes.

**With mountineers**, what was formerly a separate action routine run in a
separate child process, can now be done *directly in-process*, simplifying
program design (and logging as well as debugging) significantly. No more passing
potentially large chunks of information forth and especially back between parent
and child processes. The key here is a particular architectural feature of Linux
and its process filesystem in particular: the `root` elements for each process
in the process filesystem, such as `/proc/1/root`.

Given proper capabilities (namely, `CAP_SYS_PTRACE`), these `root` elements form
"wormholes" into the filesystem as seen by its process, subject to `chroot` and
... mount namespaces!

In case there is already a process attached to a mount namespace we want to read
from, we can simply address the mount namespace filesystem view through the
process filesystem entry for the attached process.

And if there's no such attached process, we simply create one ourselves and keep
it alive, albeit completely idle, while we need to access the mount namespace's
filesystem view. Here, lxkns either uses a dedicated pause binary (see
`cmd/mntnssandbox`) or, if that binary cannot be found, re-executes itself in
order to turn this child copy into an idle pause process.

The separate `mntnssandbox` binary is preferable from the perspective of
minimizing resource consumption, especially when re-executing a large binary.

#### Notes

[^1]: not without reason, the good people of
      [Franconia](https://en.wikipedia.org/wiki/Franconia) in Germany are also
      known as "consonant defilers". They defile not only German g's, k's, b's
      and p's, but also English th's.
