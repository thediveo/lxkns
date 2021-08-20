# Laboratory Notes

Things that truely don't belong into the information model section.

## Namespace Discovery

- function `Discover()`
- type `DiscoveryResult`

### Capabilities

Discovering namespaces from the `proc` filesystem (usually mounted as `/proc`)
requires sufficient "privileges" – capabilities – in order to query processes
owned by users different than the one running a `lxkns` namespace discovery.
More specific, this requires `CAP_SYS_PTRACE`, which is a rather nasty
capability, allowing to stop and trace other processes, to dump them, and many
more god-like things like shamelessly peeking into the mount namespaces of
unsuspecting containers.

### Multiple "Root" PID and user Namespaces

When a discovery is run without sufficient privileges, it might yield slightly
"unintuitive" results. In particular, when PID and user namespaces are
bind-mounted and/or fd-references, these can still be found despite the `lxkns`
discovery process not having sufficient privileges to glance them from the
`proc` filesystem (though still subject to having access to the bind-mounts and
fd references).

However, usually privileges are insufficient to find the parent namespace of
such PID and user namespaces found in unexpected corners, so
`relations.Parent()` as our facade to the corresponding Linux `ioctl()` only
gives us `access denied` … and this doesn't distinguish between "no access" and
"no parent", probably in an attempt of information hiding, née "sekuriti".

In consequence, `lxkns` will then return multiple PID and/or user namespace
"roots" in `DiscoveryResult.UserNSRoots` and `DiscoveryResult.PIDNSRoots`.

## Process Discovery

- type `ProcessTable`
- factory `NewProcessTable()`

Nota bene:

Discovering the process hierarchy and process status doesn't need special
privileges (capabilities). In particular, free access is given to:
  
- `/proc/[PID]/stat` with PID, PPID, status, ...
- `/proc/[PID]/status` with list of PIDs in current and parent PID namespaces.

However, discovering the Linux kernel namespaces to which the processes are
joined to **requires privileges**, and the `CAP_SYS_PTRACE` capability in
particular. `/proc/[PID]/ns/*` are accessible only to processes with
`CAP_SYS_PTRACE`. This can be checked using the following command which should
succeed, unless you remove the `capsh` CLI argument `--addamb=cap_sys_ptrace` to
make it fail with permission denied:

```bash
capsh --caps="cap_sys_ptrace+eip cap_setpcap,cap_setuid,cap_setgid+ep" \
    --keep=1 --user=nobody --addamb=cap_sys_ptrace \
    -- -c "ls -l /proc/1/ns/net"
```

Please note that this command requires a reasonable recent version of `capsh`
which supports ambient capabilities. It almost goes without saying, that once
more Debian and Raspbian don't fit the bill. You'll need to compile `capsh` from
the [libcap sources](https://git.kernel.org/pub/scm/libs/libcap/libcap.git/) on
such “stable” broken distributions.

- [How do I use capsh: I am trying to run an unprivileged ping, with minimal
  capabilities](https://unix.stackexchange.com/a/303738) (stackexchange)
- [Access /proc/pid/ns/net without running query process as
  root?](https://unix.stackexchange.com/a/561106) (stackexchange)

This has ramifications especially to the `pidtree` command (package
`./cmd/pidtree`): when run with insufficient privileges (capabilities), then:

- when run in the initial PID namespace, quite some process hierarchy above the
  `pidtree` process will lack their namespace-related information. In
  consequence, it makes more sense to only start with `pidtree`'s PID namespace
  and its "leader" processes (that is, all topmost processes still "inside" the
  PID namespace).

  - `pidtree`'s PID namespace might be the initial PID namespace, but we cannot
    know – even when we're able to access PID 1, we might still be inside a
    parent PID namespace … and we simply cannot detect this, as this is
    exactly the way how Linux kernel PID namespaces are designed to work.
  
  - we might discover other bind-mounted or fd-referenced PID namespaces, which
    seem to have **no parent** PID namespace, so they appear to be "root
    namespaces". Please see above for further discussion of this situation.
