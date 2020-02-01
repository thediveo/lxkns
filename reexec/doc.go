/*

Package reexec allows to fork and then re-execute the current application
(process) in order to only invoke a specific action function.

Why, because the Golang runtime sucks at fork() and switching Linux kernel
mount namespaces. The go runtime spins up multiple threads, but Linux really
doesn't like changing mount namespaces when a process has become
multi-threaded.

For switching Linux kernel namespaces on re-execution, we rely on the package
github.com/thediveo/gons. It strictly requires cgo, because it needs to run
plain C code for switching namespaces and be done before the go runtim spins
up.

*/
package reexec
