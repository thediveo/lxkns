/*

Package species defines the type constants and type names of the 7 Linux kernel
namespace types ("species"). In addition, this package converts between the type
names and their corresponding (Linux kernel) constants.

While Golang's x/sys/unix package finally defines even the formerly missing
CLONE_NEWCGROUP constant, this package redefines the namespace-related
CLONE_NEWxxx identifiers to be type-safe, so they cannot accidentally be mixed
with other CLONE_xxx constants.

*/
package species
