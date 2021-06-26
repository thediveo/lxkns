# Namespace Discovery

**lxkns** was born from the need for a generic discovery API and service for
Linux-kernel namespaces, to build various kinds of system diagnosis and
performance tools upon. A need for discovery that starts at the Linux-kernel API
level and optionally takes containers into consideration, without being tied to
a specific container engine.

## Discovery Range

**lxkns** features an "extended discovery range", when compared to most
well-known and openly available CLI tools and especially the golden oldie
`lsns`. lxkns detects namespaces even in places of a running Linux system other
tools typically do not consider.

The following places are searched for traces of namespaces:

1. of course, from the **procfs** filesystem in `/proc/[PID]/ns/*` – as `lsns`
   and all the other namespace-related tools do.
2. **bind-mounted namespaces**, via `/proc/[PID]/mountinfo`. lxkns even takes
   bind-mounted namespaces in _other_ mount namespaces than the current/initial
   mount namespace into account.
3. **fd-referenced namespaces**, via `/proc/[PID]/fd/*`.
4. **intermediate hierarchical user and PID namespaces**, via `NS_GET_PARENT`
   ([man 2 ioctl_ns](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).
5. **user namespaces owning non-user namespaces**, via `NS_GET_USERNS` ([man 2
   ioctl_ns](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).

Or in table format:

| tool | `/proc/[PID]/ns/*` ① | bind mounts ② | `/proc/[PID]/fd/*` ③ | hierarchy ④ | owning user namespaces ⑤ |
| --- | --- | --- | --- | --- | --- |
| `lsns` | ✓ | | | |
| `lxkns` | ✓ | ✓ | ✓ | ✓ | ✓ |

Applications can control the extent to which a `lxkns` discovery tries to
ferret out namespaces from the nooks and crannies of Linux hosts.

> [!NOTE] Some discovery methods are more expensive than others, especially the
> discovery of bind-mounted namespaces in other mount namespaces. The reason
> lies in the design of the Go runtime which runs multiple threads and Linux not
> allowing multi-threaded processes to switch mount namespaces. In order to work
> around this constraint, `lxkns` must fork and immediately re-execute the
> process it is used in. Applications that want to use such advanced discovery
> methods thus **must** call `reexec.CheckAction()` as early as possible in
> their `main()` function. For this, you need to `import
> "github.com/thediveo/gons/reexec"`.

## Required Capabilities

**lxkns** discoveries require the following capabilities:

- `CAP_SYS_PTRACE` grants access to the namespace information in the `/proc`
  filesystem.

- `CAP_SYS_ADMIN` and `CAP_SYS_ADMIN` grants switching into other (especially
  mount) namespaces in order to look into more places compared to standard
  discovery tools. Additionally, these capabilities allows a discovery service
  to switch back into the initial cgroup namespace in order to discover correct
  cgroup hierarchy information. Similarly, temporarily switching into the
  initial mount namespace allows us to correctly pick up the freezer ("fridge")
  states of processes, this works around having to either explicitly mount the
  host's cgroup into the container or to unprotect the container's system paths
  (which docker-compose yet does not support).

- `CAP_DAC_READ_SEARCH` grants discovering bind-mounted namespaces without
  interference by any DAC, or "(in)descretionary axcess control".
