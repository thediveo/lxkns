/*

Package species defines the type constants and type names of the 7 Linux kernel
namespace types ("species"). In addition, this package also defines how to
represent namespace identifiers, which actually consist of not only an inode
number, but also the device ID where a namespace inode is located on. This
package also converts between the namespace type names and their corresponding
(Linux kernel) constants, as well as between the internal and textual
representations of namespace identifiers.

Namespace Identifiers

Caveat: currently, the textual representation of namespace identifiers employed
by the Linux kernel and CLI tools ignores the device ID part of a complete
namespace identifier.

Namespace Type Constants

While Golang's x/sys/unix package finally defines even the formerly missing
CLONE_NEWCGROUP constant (which was missing from the syscall package), this
package still redefines the namespace-related CLONE_NEWxxx identifiers to be
type-safe. This way, they cannot accidentally be mixed with other CLONE_xxx
constants, or the CLONE_xxx flags in general.

*/
package species
