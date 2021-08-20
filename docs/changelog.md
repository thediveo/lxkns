# Important Changes

## 0.19.0

1. [`model.Namespace.Ref()`](https://pkg.go.dev/github.com/thediveo/lxkns/model#Namespace)
   now returns a slice of strings, instead of a single string. The rationale is
   that bind-mounted namespaces might be bound in mount namespaces other than
   the initial mount namespace. The updated API now correctly reflects this.

2. It is not necessary anymore to call `reexec.CheckAction()` as soon as
   possible in any application importing `lxkns`: please remove all imports and
   calls to `github.com/thediveo/gons/reexec`. The new implementation uses less
   system resources and makes collecting coverage information much easier and
   reliable.

   <div class="backgroundinfo">

   The discovery engine has been reworked and simplified in order to now
   directly access file system contents inside (other) mount namespaces via the
   process file system (via the `root` elements). In most cases, this avoids the
   need to spin up a new process that then is used to read from another mount
   namespace with inter-process communication and all the shebang. Instead,
   lxkns can now **read directly** from another mount namespace via the process
   file system without any need for fiddly inter-process communication.

   In the remaining rare cases – typically bind-mount mount namespaces – lxkns
   still spins up a new process, but now doesn't need to channel any file system
   accesses through it. Instead, the new process immediately goes to sleep after
   successfully switching to the desired mount namespace and later gets killed
   when it isn't needed any longer. All lxkns needs, is a suitable entry in the
   process file system, so a sleeper sufficies. Please see also
   [mntnssandbox](mntnssandbox).

   The exact details are hidden away in an easy-to-use type, mockingly called a
   "[mountineer](mountineers)": it mounts mount namespaces, albeit not in the
   usual sense of mounting a file system using the `mount` syscall.

   </div>
