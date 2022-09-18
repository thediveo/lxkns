/*
pidtree displays a tree (or only a single branch) of processes together with
their PID namespaces, and additionally also shows the local PIDs of processes
(where applicable).

The pidtree command bears some resemblance with the pstree command
(http://man7.org/linux/man-pages/man1/pstree.1.html) in that both display a tree
of processes. However, pidtree focuses on how processes are organized in PID
namespaces; something pstree isn't aware of.

# Usage

To use pidtree:

	pidtree [flag]

For example, to view the colorized tree of PID namespaces and their processes in
a pager:

	pidtree -c | less -SR

# Flags

The following pidtree flags are available:

	    --all-leaders            show all leader processes instead of only the most senior one
	    --cgroup cgformat        control group name display; can be 'full' or 'short' (default short)
	-c, --color color[=always]   colorize the output; can be 'always' (default if omitted), 'auto',
	                             or 'never' (default auto)
	    --containerd string      containerd engine API socket path (default "/run/containerd/containerd.sock")
	    --docker string          Docker engine API socket path (default "unix:///var/run/docker.sock")
	    --dump                   dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)
	-h, --help                   help for pidtree
	    --icon                   show/hide unicode icons next to namespaces
	    --nocontainerd           do not consult a containerd engine
	    --nodocker               do not consult a Docker engine
	    --noengines              do not consult any container engines
	-n, --ns string              PID namespace of PID, if not the initial PID namespace;
	                             either an unsigned int64 value, such as "4026531836", or a
	                             PID namespace textual representation like "pid:[4026531836]"
	-p, --pid uint32             PID of process to show PID namespace tree and parent PIDs for
	    --proc proc[=name]       process name style; can be 'name' (default if omitted), 'basename',
	                             or 'exe' (default name)
	    --theme theme            colorization theme 'dark' or 'light' (default dark)
	    --treestyle treestyle    select the tree render style; can be 'line' or 'ascii' (default line)
	-v, --version                version for pidtree

# Display

The process tree starts at the topmost PID namespace; when started in the
initial PID namespace, then this will be the initial PID namespace.

	pid:[4026531836], owned by UID 0 ("root")
	├─ "systemd" (1748)
	[...]

When started in a child PID namespace, then the topmost PID namespace will be
that child PID namespace. The Linux kernel makes it impossible for processes to
reach out into parent (or sibling) PID namespaces, thus pidtree cannot show
proper PID namespacing information for such processes above the starting point's
PID namespace (please also see below).

Whenever a child process lives in a different PID namespace than its parent
process, pstree shows an intermediate PID namespace node between parent and
child process(es). These PID namespace nodes show the namespace ID (inode
number), as well as the user ID and user name "owning" the PID namespace. For a
PID namespace, the owner is the user which created the user namespace, which in
turn was active when the PID namespace was created. Or to phrase this chain
slightly differently: the PID namespace is owned by a user namespace, and that
user namespace is owned by a user.

	pid:[4026531836], owned by UID 0 ("root")
	├─ "systemd" (1)
	│  ├─ "systemd-journal" (417)
	[...]
	│  │  └─ "unshare" (5309)
	│  │     └─ pid:[4026532229], owned by UID 1000 ("thediveo")
	│  │        └─ "bash" (5310/1)
	│  │           └─ "unshare" (5344/24)
	│  │              └─ pid:[4026532247], owned by UID 1000 ("thediveo")
	│  │                 └─ "bash" (5345/1)
	│  │ └─ "sleep" (5529/25)
	[...]

For PID namespaces different from the starting PID namespace, pidtree not only
shows the process PIDs as seen from the starting PID namespace. Additionally, it
also shows the PID from the perspective of the "local" PID namespace. The
"local" PID namespace is the PID namespace a process is joined to. The "local"
PID is thus the PID seen by the process itself. For instance, this information
about the local PID might be useful when entering child PID namespaces and
controlling childs or diagnosing logs with PIDs.

	[...]
	│  │              └─ pid:[4026532247], owned by UID 1000 ("thediveo")
	│  │                 └─ "bash" (5345/1)
	│  │                    └─ "sleep" (5529/25)
	[...]

Insufficient Privileges/Capabilities:

When pidtree is started without the necessary privileges (in particular, the
CAP_SYS_PTRACE capability) it has only limited visibilty onto Linux kernel
namespaces. In this situation, pidtree will only show processes which are either
in the same PID namespace as itself, or which are direct or indirect children of
such processes. When a child process is in an inaccessible PID namespace, then
it will be prefixed with a "pid:[???]" indication and additionally an unknown
local PID "/???" will be shown.

	[...]
	│  ├─ "bash" (14725)
	│  │  └─ pid:[???] "sudo" (14738/???)
	│  │     └─ pid:[???] "unshare" (14742/???)
	│  │        └─ pid:[???] "bash" (14744/???)
	│  │           └─ pid:[???] "unshare" (14756/???)
	│  │              └─ pid:[???] "bash" (14757/???)
	│  │                 ├─ pid:[???] "sleep" (14773/???)
	│  │                 └─ pid:[???] "bash" (15662/???)
	[...]

# Colorization

Unless specified otherwise using the "--color=none" flag, pidtree colorizes its
output in order to make different types of namespaces easier to differentiate.
Colorization gets disabled if pidtree detects that stdout isn't connected to a
terminal, such as when piping into tools like "less".

Out of the box (or rather, Gopher hole), pidtree supports two color themes,
called "dark" and "light". Default is the dark theme, but it can be changed
using "--theme light". In order to set a theme permanently, and to optionally
adapt it later to personal preferences, the selected theme can be written to
stdout:

	pidtree --theme light --dump > ~/.lxknsrc.yaml

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
