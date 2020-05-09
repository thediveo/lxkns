/*

nscaps determines a process capabilities in some namespace. It shows the process
and namespace in their namespace hierarchy or hierarchies, together with the
process' own effective capabilities as well as the capabilities set gained in
the specific namespace.

The nscaps tool has been inspired by Michael Kerrisk's gem
http://man7.org/tlpi/code/online/dist/namespaces/ns_capable.c, but is an
independent implementation.

Usage

To use nscaps:

    nscaps [flag] NAMESPACE

For example, to view the capabilities of the current shell in your current
network namespace (which are usually none as an ordinary user):

    nscaps $(readlink /proc/$$/ns/net)

Compare this with root gets:

    sudo nscaps $(readlink /proc/$$/ns/net)

Flags

The following nscaps flags are available:

        --brief                  show only a summary statement for the capabilities in the target namespace
    -c, --color color[=always]   colorize the output; can be 'always' (default if omitted), 'auto',
                                 or 'never' (default auto)
        --dump                   dump colorization theme to stdout (for saving to ~/.lxknsrc.yaml)
    -h, --help                   help for nscaps
        --icon                   show/hide unicode icons next to namespaces
    -n, --ns string              PID namespace of PID, if not the initial PID namespace;
                                 either an unsigned int64 value, such as "4026531836", or a
                                 PID namespace textual representation like "pid:[4026531836]"
    -p, --pid uint32             PID of process for which to calculate capabilities
        --proc proc[=name]       process name style; can be 'name' (default if omitted), 'basename',
                                 or 'exe' (default name)
        --proccaps               show the process' capabilities (default true)
        --theme theme            colorization theme 'dark' or 'light' (default dark)
        --treestyle treestyle    select the tree render style; can be 'line' (default if omitted)
                                 or 'ascii' (default line)
    -v, --version                version for nscaps

Display

nscaps shows a tree containing only one or at most two branches: one containing
the specified process (defaults to the nscaps process), and a second containing
the specified namespace.

    user:[4026531837] process "systemd" (129419)
    ├─ process "nscaps" (208439)
    │     ⋄─ (no capabilities)
    └─ target net:[4026531905] process "systemd" (129419)

For instance, a process inside a user-created user namespace (sic!) gains (all)
capabilities ... how shockingly shocking! In the followin example, the process
is a user process inside the initial userspace and the target namespace is a
network namespace inside the user-created user namespace (sic!). As user
namespaces are designed, this grants the process all capabilities in the network
namespace.

    user:[4026531837] process "systemd" (129419)
    ├─ process "nscaps" (210373)
    │     ⋄─ (no capabilities)
    └─ user:[4026532342] process "unshare" (176744)
       └─ target net:[4026532353] process "unshare" (176744)
             ⋄─ cap_chown            cap_dac_override     cap_dac_read_search  cap_fowner
             ⋄─ cap_fsetid           cap_kill             cap_setgid           cap_setuid
             ⋄─ cap_setpcap          cap_linux_immutable  cap_net_bind_service cap_net_broadcast
             ⋄─ cap_net_admin        cap_net_raw          cap_ipc_lock         cap_ipc_owner
             ⋄─ cap_sys_module       cap_sys_rawio        cap_sys_chroot       cap_sys_ptrace
             ⋄─ cap_sys_pacct        cap_sys_admin        cap_sys_boot         cap_sys_nice
             ⋄─ cap_sys_resource     cap_sys_time         cap_sys_tty_config   cap_mknod
             ⋄─ cap_lease            cap_audit_write      cap_audit_control    cap_setfcap
             ⋄─ cap_mac_override     cap_mac_admin        cap_syslog           cap_wake_alarm
             ⋄─ cap_block_suspend    cap_audit_read

Insufficient Privileges/Capabilities:

When nscaps is started without the necessary privileges (in particular, the
CAP_SYS_PTRACE capability) it has only limited visibilty onto Linux kernel
namespaces. In this situation nscaps cannot determine the capabilities when
either the specified process or specified namespace is totally out of reach from
nscap's perspective.

Colorization

Unless specified otherwise using the "--color=none" flag, nscaps colorizes its
output in order to make different types of namespaces easier to differentiate.
Colorization gets disabled if nscaps detects that stdout isn't connected to a
terminal, such as when piping into tools like "less".

Out of the box (or rather, Gopher hole), nscaps supports two color themes,
called "dark" and "light". Default is the dark theme, but it can be changed
using "--theme light". In order to set a theme permanently, and to optionally
adapt it later to personal preferences, the selected theme can be written to
stdout:

    nscaps --theme light --dump > ~/.lxknsrc.yaml

For each type of Linux-kernel namespace the styling file "~.lxknsrc.yaml"
contains a top-level element:

    user:
    pid:
    cgroup:
    ipc:
    mnt:
    net:
    uts:

Additional output elements can also be styled:

    process: # process names
    owner:   # owner UIDs and user names
    unknown: # unknown PIDs and PID namespaces

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
