# Important Changes

## 0.20.0

1. lxkns has grown tremendously in the past more than one-and-a-half year since
   its first steps. Not unexpectedly, this has lead into the issue of the main
   `github.com/thediveo/lxkns` repo folder to hoard lots of discovery-related
   source files, despite all the other sub packages. For this reason, almost all
   discovery-related stuff has been moved into its own `discovery` package, out
   of `lxkns`. In order to avoid unnecessary stuttering, a few types have been
   renamed too.
   - rename all imports of `github.com/thediveo/lxkns` to
     `github.com/thediveo/lxkns/discover`.
   - rename all uses of `lxkns.Discover` to `discover.Namespaces`.
   - rename all uses of `lxkns.DiscoveryResult` to `discover.Result`

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
