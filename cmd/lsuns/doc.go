/*
lsuns lists the tree of user namespaces, optionally with the other namespaces
they own.

# Usage

To use lsuns:

	lsuns [flag]

For example, to view the colorized tree of user and owned namespaces in a pager:

	lsuns -cd | less -SR

# Flags

The following lsuns flags are available:

	    --all-leaders            show all leader processes instead of only the most senior one
	    --cgroup cgformat        control group name display; can be 'full' or 'short' (default short)
	-c, --color color[=always]   colorize the output; can be 'always' (default if omitted), 'auto',
	                             or 'never' (default auto)
	-d, --details                shows details, such as owned namespaces
	    --dump                   dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)
	-f, --filter filter          shows only selected namespace types; can be 'cgroup'/'c', 'ipc'/'i', 'mnt'/'m',
	                             'net'/'n', 'pid'/'p', 'user'/'U', 'uts'/'u' (default [mnt,cgroup,uts,ipc,user,pid,net])
	-h, --help                   help for lsuns
	    --icon                   show/hide unicode icons next to namespaces
	    --proc proc[=name]       process name style; can be 'name' (default if omitted), 'basename',
	                             or 'exe' (default name)
	    --theme theme            colorization theme 'dark' or 'light' (default dark)
	    --treestyle treestyle    select the tree render style; can be 'line' or 'ascii' (default line)
	-v, --version                version for lsuns
	    --wait duration          max duration to wait for container engine workload synchronization (default 3s)

# Colorization

Unless specified otherwise using the "--color=none" flag, lsuns colorizes its
output in order to make different types of namespaces easier to differentiate.
Colorization gets disabled if lsuns detects that stdout isn't connected to a
terminal, such as when piping into tools like "less".

Out of the box (or rather, Gopher hole), lsuns supports two color themes, called
"dark" and "light". Default is the dark theme, but it can be changed using
"--theme light". In order to set a theme permanently, and to optionally adapt it
later to personal preferences, the selected theme can be written to stdout:

	lsuns --theme light --dump > ~/.lxknsrc.yaml

For each type of Linux-kernel namespace the styling file "~.lxknsrc.yaml"
contains a top-level element:

  - user:
  - pid:
  - cgroup:
  - ipc:
  - mnt:
  - net:
  - uts:

Additional output elements can also be styled:

  - process: # process names
  - owner:   # owner UIDs and user names
  - unknown: # unknown PIDs and PID namespaces

For each top-level element the foreground and background colors can be set
independently, as well as several different type face and font rendering
attributes. If the foreground and/or background color(s) or a specific attribute
are not specified, then the terminal defaults apply.

Colors and attributes need to be specified in form of YAML list members,
introduced with a "-" dash. Colors can be specified either in #RRGGBB format, or
alternatively as ANSI colors (0-255). Make sure to always enclose color values
in (single or double) quotes.

For example:

	pid:
	- bold
	- foreground: '#aabbcc'

The following attributes are supported, but are subject to specific terminal
implementations rendering them:

  - blink
  - bold
  - crossout
  - faint
  - italic
  - italics
  - overline
  - reverse
  - underline
*/
package main
