/*

Package nstypes defines the type constants and type names of the 7 Linux
kernel namespaces. In addition, it converts between these type names and their
corresponding constants.

Unfortunately, Go's syscall package lacks the constant definition for
CLONE_NEWCGROUP. In consequence, we have to define the full set of namespace
type constants anyway, and then can also properly type these constants to be
NamespaceType as an added bonus.

*/
package nstypes
