/*

Package portable provides so-called "portable" namespace references with
validation and "locking" (keeping the referenced namespace open and thus alive).

There's an unavoidable non-zero timing window between the discovery of
namespaces (and their references) and attempting to use them for switching
namespaces. In this window, namespace references might become invalid without
noticing it at first: when using a namespace ID or a bind-mounted namespace
path, in the worst case the old namespace might have been garbage collected, yet
the reference might now point to a different namespace with a recycled
reference.

To correctly detect such unwanted situations, more namespace-related data is
needed to thoroughly cross-check a namespace reference before using it. This
cross-check information is represented in form of PortableReference objects. A
PortableReference additionally ensures that the namespace cannot change anymore
between cross-checking it and using it, for instance with ops.Enter().

Cross-checking, or validation, is done together with "locking" the namespace in
one integral step by calling Open() on a given PortableReference. In the
following example we've left out proper error checking for sake of brevity:

    portref := portable.PortableReference{
        ID: 4026531837,
        Type: species.CLONE_NEWNET,
        PID: 12345,
        StartingTime: 1234567890,
    }
    lockedref, unlocker, _ := portref.Open()
    defer unlocker()
    ops.Execute(
        func()interface{}{
            // do something useful in here while attached to the network namespace...
            return nil
        },
        lockedref)

Important

Make sure that the returned lockedref doesn't get garbage collected too early.
In the above example this is ensured by ops.Executing getting the lockedref
passed. Depending on your specific use case you might need to place a
runtime.KeepAlive(lockedref) beyond the point where you definitely need the
lockedref to be still correctly locked.

Note

As with switching namespaces in Go applications in general, please remember that
it is not possible to switch the mount and user namespaces in OS multi-threaded
applications.

*/
package portable
