/*

Package mntnssandbox is a single-purpose, stripped-down version of
thediveo/gons. Our variant here only supports switching the mount and user
namespaces, and then immediately goes into endless sleep. There's neither
switching multiple other namespaces nor running any registered action. All there
is is endless sleep.

Why should this be useful at all? For accessing the contents of a mount
namespace (target mount namespace) different from the mount namespace of the
accessing process a neat trick is to actually not switch the accessor process
into the target mount namespace. Instead, it is much simpler to just switch a
"most simple" process into that mount namespace and then let that process sleep.
While the process sleeps the original access process now can access the target
mount namespace via the proc file system: /proc/[SLEEPINGPID]/root.

The benefits are twofold: first, we need only a tiny sleeper process consuming
few resources, as compared to having to fork and re-execute our own huge binary
to enter the target mount namespace. Second, we don't need to marshal and
unmarshal the information passed forth and back between the original process and
the child process. All we do is simply directly accessing the file system
contents from the original process, while the sleeper process sleeps.

This package gets executed automatically during startup and before the Go
runtime spins up. If there is no environment variable "sleepy_mntns" set (or it
has an empty value), then further initialization will proceed as usual and this
package will keep neutral.

However, if the environment variable "sleepy_mntns" has been set with a
non-empty value, then it is a file system pathname referencing a mount
namespace. If switching into the specified mount namespace fails, then an error
message will be sent to fd 2 (stderr) and the program aborted. If switching
succeeds, then an "OK" message will be sent to fd 1 (stdout) and the program
will block indefinitely, keeping the referenced mount namespace accessible via
the proc file system, and this process in particular.

If an additional environment variable "sleepy_userns" has been specified then
the referenced user mount namespace will be entered first before entering the
mount namespace. This allows entering child user namespaces belonging to one's
own user without the otherwise required effective nameswitching capabilities
(CAP_SYS_ADMIN and CAP_SYS_CHROOT in case of mount namespaces).

*/
package mntnssandbox
