# CLI Tools

## Installation

### System Installation

```bash
make install
```

…builds and install all CLI tools into your system, defaults to
`/usr/local/bin`.

### Local Installation

```bash
go install ./cmd/... ./examples/lsallns
```

…installs all CLI tools into `$GOPATH/bin`.

## Colorful Results

Most lxkns CLI tools support colorizing their output to aid in spotting and
differentiating the various types of namespaces.

- `-c`, `--color` and `--color always`: unconditionally colorizes discovery
  results, whether the output is sent to a terminal or a pipe, file, et cetera.
- `--color auto`: colorizes results only if sending to a terminal.
- `--color never`: never colorizes results, event if sending to a terminal.

## lsuns

In its simplest form, `lsuns` shows the hierarchy of user namespaces.

```console
$ sudo lsuns
user:[4026531837] process "systemd" (1) created by UID 0 ("root")
├─ user:[4026532454] process "unshare" (98171) controlled by "user.slice" created by UID 1000 ("harald")
└─ user:[4026532517] process "upowerd" (96159) controlled by "system.slice/upower.service" created by UID 0 ("root")
```

> [!NOTE] `lsuns` does not only show the user namespaces with their IDs and
> hierarchy. It also shows the "most senior" process attached to the particular
> user namespace, as well as the user "owning" the user namespace. The "most
> senior" process is the top-most process in the process tree still attached to
> a (user) namespace; in case of multiple top-most processes – such as init(1)
> and kthreadd(2) – the older process will be choosen (or the one if the lowest
> PID as in case of the same-age init and kthreadd).

The control group name ("controlled by ...") is the name of the v1 "cpu" control
sub-group controlling a particular most senior process. This name is relative to
the root of the control group filesystem (such as `/sys/fs/cgroup`). The root is
left out in order to reduce clutter.

### Showing Owned (Non-User) Namespaces

It gets more interesting with the `-d` (details) flag: `lsuns` then additionally
displays all non-user namespaces owned by the user namespaces. In Linux-kernel
namespace parlance, "owning" refers to the relationship between a newly created
namespace and the user namespace that was active at the time the new namespace
was created. For convenience, `lsuns` sorts the owned namespaces first
alphabetically by type, and second numerically by namespace IDs.

```console
$ sudo lsuns -d
user:[4026531837] process "systemd" (1) created by UID 0 ("root")
│  ⋄─ cgroup:[4026531835] process "systemd" (1)
│  ⋄─ ipc:[4026531839] process "systemd" (1)
│  ⋄─ ipc:[4026532332] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  ⋄─ ipc:[4026532397] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
│  ⋄─ mnt:[4026531840] process "systemd" (1)
│  ⋄─ mnt:[4026531860] process "kdevtmpfs" (33)
│  ⋄─ mnt:[4026532184] process "systemd-udevd" (946) controlled by "system.slice/systemd-udevd.service"
│  ⋄─ mnt:[4026532245] process "haveged" (1688) controlled by "system.slice/haveged.service"
│  ⋄─ mnt:[4026532246] process "systemd-timesyn" (1689) controlled by "system.slice/systemd-timesyncd.service"
│  ⋄─ mnt:[4026532248] process "systemd-network" (1709) controlled by "system.slice/systemd-networkd.service"
│  ⋄─ mnt:[4026532267] process "systemd-resolve" (1711) controlled by "system.slice/systemd-resolved.service"
│  ⋄─ mnt:[4026532268] process "NetworkManager" (1757) controlled by "system.slice/NetworkManager.service"
│  ⋄─ mnt:[4026532269] bind-mounted at "/run/snapd/ns/lxd.mnt"
│  ⋄─ mnt:[4026532325] process "irqbalance" (1761) controlled by "system.slice/irqbalance.service"
│  ⋄─ mnt:[4026532326] process "systemd-logind" (1779) controlled by "system.slice/systemd-logind.service"
│  ⋄─ mnt:[4026532327] process "ModemManager" (1840) controlled by "system.slice/ModemManager.service"
│  ⋄─ mnt:[4026532330] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  ⋄─ mnt:[4026532388] process "bluetoothd" (2239) controlled by "system.slice/bluetooth.service"
│  ⋄─ mnt:[4026532395] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
│  ⋄─ mnt:[4026532513] process "colord" (83614) controlled by "system.slice/colord.service"
│  ⋄─ mnt:[4026532516] process "upowerd" (96159) controlled by "system.slice/upower.service"
│  ⋄─ net:[4026531905] process "systemd" (1)
│  ⋄─ net:[4026532191] process "haveged" (1688) controlled by "system.slice/haveged.service"
│  ⋄─ net:[4026532274] process "rtkit-daemon" (2211) controlled by "system.slice/rtkit-daemon.service"
│  ⋄─ net:[4026532335] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  ⋄─ net:[4026532400] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
│  ⋄─ pid:[4026531836] process "systemd" (1)
│  ⋄─ pid:[4026532333] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  ⋄─ pid:[4026532398] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
│  ⋄─ uts:[4026531838] process "systemd" (1)
│  ⋄─ uts:[4026532185] process "systemd-udevd" (946) controlled by "system.slice/systemd-udevd.service"
│  ⋄─ uts:[4026532247] process "systemd-timesyn" (1689) controlled by "system.slice/systemd-timesyncd.service"
│  ⋄─ uts:[4026532324] process "systemd-logind" (1779) controlled by "system.slice/systemd-logind.service"
│  ⋄─ uts:[4026532331] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  ⋄─ uts:[4026532396] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
├─ user:[4026532454] process "unshare" (98171) controlled by "user.slice" created by UID 1000 ("harald")
│     ⋄─ mnt:[4026532455] process "unshare" (98171) controlled by "user.slice"
│     ⋄─ mnt:[4026532457] process "unshare" (98172) controlled by "user.slice"
│     ⋄─ pid:[4026532456] process "unshare" (98172) controlled by "user.slice"
│     ⋄─ pid:[4026532458] process "bash" (98173) controlled by "user.slice"
└─ user:[4026532517] process "upowerd" (96159) controlled by "system.slice/upower.service" created by UID 0 ("root")
```

## lspidns

On its surface, `lspidns` might appear to be `lsuns` twin, but now for PID namespaces.

```console
pid:[4026531836] process "systemd" (1)
├─ pid:[4026532333] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  └─ pid:[4026532398] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
└─ pid:[4026532456] process "unshare" (99577) controlled by "user.slice"
   └─ pid:[4026532459] process "unshare" (99578) controlled by "user.slice"
      └─ pid:[4026532460] process "bash" (99579) controlled by "user.slice"
```

> [!NOTE] If you look closely at the control group names of the PID namespace
> processes, then you might notice that there is an outer Docker container with
> an inner container. This inner container happens to be a containerd container
> in the "default" namespace.

### User~PID Namespaces Hierarchy

Hidden beneath the surface lies the `-u` flag: "u" as in user namespace.

Now what have user namespaces to do with PID namespaces? Like all other non-user
namespaces, also PID namespaces are *owned* by user namespaces. `-u` now tells
`lspidns` to show a "synthesized" hierarchy where owning user namespaces and
owned PID namespaces are laid out in a single tree.

```console
user:[4026531837] process "systemd" (1) created by UID 0 ("root")
└─ pid:[4026531836] process "systemd" (1)
   ├─ pid:[4026532333] process "systemd" (5492) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
   │  └─ pid:[4026532398] process "sleep" (6025) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
   └─ user:[4026532454] process "unshare" (99576) controlled by "user.slice" created by UID 1000 ("harald")
      └─ pid:[4026532456] process "unshare" (99577) controlled by "user.slice"
         └─ user:[4026532457] process "unshare" (99577) controlled by "user.slice" created by UID 1000 ("harald")
            └─ pid:[4026532459] process "unshare" (99578) controlled by "user.slice"
               └─ pid:[4026532460] process "bash" (99579) controlled by "user.slice"
```

> [!NOTE] This tree-like representation is possible because the capabilities
> rules for user and PID namespaces forbid user namespaces criss-crossing PID
> namespaces and vice versa. If criss-crossing would be possible, there would be
> no way to represent the information as a tree.

## pidtree

`pidtree` shows either the process hierarchy within the PID namespace
hierarchy or a single branch only.

### Complete Tree

`pidtree` without further CLI flags shows the complete hierarchy of PID
namespaces. PIDs in child PID namespaces are shown not only as they are seen
within this child PID namespace, but additionally as seen from parent PID
namespace(s).

For instance, the following example shows **two separate instances** of a
`containerd` engine being deployed: **one in the host** (=initial PID namespace),
and **another one inside a container**. The containerized `containerd` instance is
shown as `"containerd" (5709/72)`: its PID is 72 inside its own PID namespace,
yet its PID is seen as 5709 from the initial PID namespace.

> [!NOTE] The namespaced PIDs of a particular process are listed in sequence
> from the initial PID namespace down to the leaf PID namespace.

```console
$ sudo pidtree
pid:[4026531836], owned by UID 0 ("root")
├─ "systemd" (1)
│  ├─ "systemd-journal" (910) controlled by "system.slice/systemd-journald.service"
│  ├─ "systemd-udevd" (946) controlled by "system.slice/systemd-udevd.service"
...
│  ├─ "containerd" (1836) controlled by "system.slice/containerd.service"
│  │  └─ "containerd-shim" (5472) controlled by "system.slice/containerd.service"
│  │     └─ pid:[4026532333], owned by UID 0 ("root")
│  │        ├─ "systemd" (5492/1) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
│  │        │  ├─ "systemd-journal" (5642/66) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/systemd-journald.service"
│  │        │  ├─ "containerd" (5709/72) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/containerd.service"
│  │        │  ├─ "setup.sh" (5712/73) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/testing.service"
│  │        │  │  └─ "ctr" (5978/107) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/testing.service"
│  │        │  └─ "containerd-shim" (5999/126) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/containerd.service"
│  │        │     └─ pid:[4026532398], owned by UID 0 ("root")
│  │        │        ├─ "sleep" (6025/146/1) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
│  │        │        └─ "sh" (6427/235/7) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
```

### Single Branch

Alternatively, it can show just a single branch down to a PID inside a
specific PID namespace.

```console
$ sudo pidtree -n pid:[4026532398] -p 7
pid:[4026531836], owned by UID 0 ("root")
└─ "systemd" (1)
   └─ "containerd" (1836) controlled by "system.slice/containerd.service"
      └─ "containerd-shim" (5472) controlled by "system.slice/containerd.service"
         └─ pid:[4026532333], owned by UID 0 ("root")
            └─ "systemd" (5492/1) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a"
               └─ "containerd-shim" (5999/126) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/system.slice/containerd.service"
                  └─ pid:[4026532398], owned by UID 0 ("root")
                     └─ "sh" (6427/235/7) controlled by "docker/c8bf69d0651425244f472e89677177e3d488274f1d242c62a50a82f35feb8c4a/default/sleepy"
```

Please see also the [pidtree
command](https://godoc.org/github.com/thediveo/lxkns/cmd/pidtree)
documentation.

## nscaps

`nscaps` calculates a process capabilities in another namespace, based on the
owning user namespace hierarchy. It then displays both the process' and target
namespace user namespace hierarchy for better visual reference how process and
target namespace relate to each other.

Examples like the one below will give unsuspecting security "experts" a series
of fits – despite this example being perfectly secure.

```console
⛛ user:[4026531837] process "systemd" (129419)
├─ process "nscaps" (210373)
│     ⋄─ (no capabilities)
└─ ✓ user:[4026532342] process "unshare" (176744)
   └─ target net:[4026532353] process "unshare" (176744)
         ⋄─ cap_audit_control    cap_audit_read       cap_audit_write      cap_block_suspend
         ⋄─ cap_chown            cap_dac_override     cap_dac_read_search  cap_fowner
         [...]
         ⋄─ cap_syslog           cap_wake_alarm
```

...it's secure, because our superpower process can't do anything outside its
realm. But the horror on the faces of security experts will be priceless.

```console
⛔ user:[4026531837] process "systemd" (211474)
├─ ⛛ user:[4026532468] process "unshare" (219837)
│  └─ process "unshare" (219837)
│        ⋄─ cap_audit_control    cap_audit_read       cap_audit_write      cap_block_suspend
│        ⋄─ cap_chown            cap_dac_override     cap_dac_read_search  cap_fowner
│        ⋄─ cap_fsetid           cap_ipc_lock         cap_ipc_owner        cap_kill
│        ⋄─ cap_lease            cap_linux_immutable  cap_mac_admin        cap_mac_override
│        ⋄─ cap_mknod            cap_net_admin        cap_net_bind_service cap_net_broadcast
│        ⋄─ cap_net_raw          cap_setfcap          cap_setgid           cap_setpcap
│        ⋄─ cap_setuid           cap_sys_admin        cap_sys_boot         cap_sys_chroot
│        ⋄─ cap_sys_module       cap_sys_nice         cap_sys_pacct        cap_sys_ptrace
│        ⋄─ cap_sys_rawio        cap_sys_resource     cap_sys_time         cap_sys_tty_config
│        ⋄─ cap_syslog           cap_wake_alarm
└─ target net:[4026531905] process "systemd" (211474)
      ⋄─ (no capabilities)
```

Please see also the [nscaps
command](https://godoc.org/github.com/thediveo/lxkns/cmd/nscaps)
documentation.

## dumpns

The lxkns namespace discovery information can also be easily made available to
your own scripts, et cetera. Without having to integrate the Go package, simply
run the `dumpns` CLI binary: it dumps fresh discovery results as JSON.

```console
$ dumpns
{
  "namespaces": {
    "4026531840": {
      "nsid": 4026531840,
      "type": "mnt",
      "owner": 4026531837,
      "reference": "/proc/2849/ns/mnt",
      "leaders": [
        2849,
        2770,
        2662,
        2847
      ]
    },
    "4026531835": {
      "nsid": 4026531835,
      "type": "cgroup",
      "owner": 4026531837,
      "reference": "/proc/2849/ns/cgroup",
      "leaders": [
        2849,
...
```
