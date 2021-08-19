/*

Package species defines the type constants and type names of the currently 8
Linux kernel namespace types ("species"). Furthermore, this package also defines
how to represent namespace identifiers: they consist of not only an inode
number, but also the device ID where a namespace inode is located on (but see
the next section below). The species package also converts between the namespace
type names (such as "mnt", "net", and so on) and their corresponding (Linux
kernel) constants (CLONE_NEWxxx), as well as between the internal and textual
representations of namespace identifiers.

Namespace Identifiers

Caveat: currently, the textual representation of namespace identifiers employed
by the Linux kernel and CLI tools ignores the device ID part of a complete
namespace identifier, but uses only the inode number.

Internally, all lxkns packages work with both the inode number as well as the
device ID of a namespace. Please see also the notes below.

Namespace Type Constants

While Golang's x/sys/unix package finally defines even the formerly missing
CLONE_NEWCGROUP constant (which was missing from the syscall package), this
package still redefines the namespace-related CLONE_NEWxxx identifiers to be
type-safe. This way, they cannot accidentally be mixed with other CLONE_xxx
constants, or the CLONE_xxx flags in general.

To provide backwards compatibility with older Go versions, namely Go 1.13, this
package defines CLONE_NEWTIME itself when there is no underlying
unix.CLONE_NEWTIME available. Applications using lxkns should thus only use
species. CLONE_NEWTIME in order to be shielded from variations in Go's unix
package.

*/
package species
