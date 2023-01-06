# Mountineers

There are Engineers, Hellsineers[^1], Mountineers...

Okay, all the joking aside, "mountineers" are used in **lxkns** to access the
file system inside those mount namespaces which are only bind-mounted and thus
currently have neither process nor task/thread attached to them. So, mountineers
mount mount namespaces (sic!) in order to access their contents using ordinary
file operations. Please note that this "mounting" **isn't** in the technical
sense of the `mount(2)` syscall, but an especially bad pun.

In fact, mountineers are used to unify access to the file system contents inside
mount namespaces in general, regardless of bind-mounted or not. The mountineers
hide all the logic to decide whether there's a convenient process already in
place that gives us access via the process filesystem, or whether we need to
spin up our own dedicated, yet temporary "sandbox" task or process.

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

Mountineers ~~work around the limitation that multi-threaded processes (and thus
Go programs) cannot switch mount namespaces~~ hide the complexity of accessing
other mount namespaces in Go programs.

> 🙏 many kudos to [Michael Kerrisk](https://www.man7.org/) for sending some
> enlightment my way about simply using
> [`unshare(2)`](https://man7.org/linux/man-pages/man2/unshare.2.html) with the
> `CLONE_FS` flag in Go.

**Before mountineers**, lxkns forked its (parent) process and then re-executed
itself in order to switch into a target mount namespace before the go runtime
spins up and called a registered "action" function (leveraging the
[thediveo/gons](https://github.com/TheDiveO/gons) module). This design has the
important drawback that optional call parameters as well as any results need to
be channelled forth and back between the parent and re-executed child processes.

**With mountineers**, what was formerly a separate action routine run in a
separate child process, can now be done often *directly in-process*, simplifying
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

And if there's no such attached process, we simply create a throw-away task (in
form of an [`LockOSThread`'ed](https://pkg.go.dev/runtime#LockOSThread)
goroutine) and then keep it alive, albeit completely idle, while we need to
access the mount namespace's filesystem view.

There's one caveat, however: in case we need to access a mount namespace owned
by a user namespace different from the one we're in, we still have to resort to
a dedicated process instead of a task/thread. The reason is that `setns()`
doesn't allow switching user namespaces in a multi-threaded process. Sigh, it
could've been so simple.

#### Notes

[^1]: not without reason, the good people of the
      [Franconia](https://en.wikipedia.org/wiki/Franconia) region in Germany are
      also known as "consonant defilers". They defile not only German g's, k's,
      b's and p's, but especially English th's.
