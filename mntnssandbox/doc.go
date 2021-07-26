/*

Package mntnssandbox is a single-purpose, stripped-down version of
thediveo/gons. Our variant here only supports switching the mount namespace and
then immediately going into endless sleep. There's neither switching multiple
namespaces nor running a registered action. All there is is endless sleep.

This package gets executed automatically during startup and before the Go
runtime spins up. If there is no environment variable "sleepy_mntns" set (or it
has an empty value), then further initialization will proceed as usual and this
package will keep neutral.

However, if the environment variable "sleepy_mntns" has been set with a
non-empty value, then it is a filesystem pathname referencing a mount namespace.
If switching into the specified mount namespace fails, then an error message
will be sent to fd 2 (stderr) and the program aborted. If switching succeeds,
then an "OK" message will be sent to fd 1 (stdout) and the program will block
indefinitely, keeping the referenced mount namespace accessible via the proc
file system, and this process in particular.

*/
package mntnssandbox
