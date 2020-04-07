/*

lsuns lists the tree of user namespaces, optionally with the other namespaces
they own.

Usage

To run lsuns, enter in CLI:

  lsuns [flag]

Flags

The following pidtree flags are available:

  -c, --color colormodus[=always]   colorize the output; can be 'always' (default if omitted), 'auto',
                                    or 'never' (default always)
  -d, --details                     shows details, such as owned namespaces
      --dump                        dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)
  -h, --help                        help for lsuns
      --theme theme                 colorization theme 'dark' or 'light' (default dark)

*/
package main
