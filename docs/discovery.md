# Namespace Discovery

**lxkns** was born from the need for a generic discovery API and service for
Linux-kernel namespaces, to build various kinds of system diagnosis and
performance tools upon. A need for discovery that starts at the Linux-kernel API
level and optionally takes containers into consideration, without being tied to
a specific container engine.

## Discovery Range

**lxkns** features an "extended discovery range", when compared to most
well-known and openly available CLI tools and especially the golden oldie
[lsns(8)](https://man7.org/linux/man-pages/man8/lsns.8.html). lxkns detects
namespaces even in places of a running Linux system other tools typically do not
consider:

| | Where? | `lsns` | `lxkns` |
| --- | --- | :---: | :---: |
| ①  | `/proc/*/ns/*` | ✓ | ✓ |
| ②  | `/proc/*/task/*/ns/*` | ✗**¹** | ✓ |
| ③  | bind mounts | ✗ | ✓ |
| ➃a | `/proc/*/fd/*` namespace fds | ✗ | ✓ |
| ➃b | `/proc/*/fd/*` socket fds | ✗ | ✓ |
| ➄  | hierarchy | ✗ | ✓ |
| ➅  | owning user namespaces | ✗ | ✓ |

1. of course, from the **procfs** filesystem in `/proc/[PID]/ns/*` – as `lsns`
   and all the other namespace-related tools do.
2. all **[tasks](https://en.wikipedia.org/wiki/Task_(computing)#Linux_kernel)**
   in `/proc/[PID]/task/[TID]/ns/*` ([Michael Kerrisk](https://www.man7.org/) of
   [The Linux Programming Interface](https://www.man7.org/tlpi/index.html) fame
   nudged me to finally fill in this gap).
3. **bind-mounted namespaces**, via `/proc/[PID]/mountinfo`. lxkns even takes
   bind-mounted namespaces in _other_ mount namespaces than the current/initial
   mount namespace into account.
4. **fd-referenced namespaces**, via `/proc/[PID]/fd/*`.
   - fd directly referencing a namespace (of any type),
   - fd referencing a socket (thus, network namespaces only).
5. **intermediate hierarchical user and PID namespaces**, via `NS_GET_PARENT`
   (for details, please refer to
   [ioctl_ns(2)](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).
6. **user namespaces owning non-user namespaces**, via `NS_GET_USERNS` (for
   details, please refer to
   [ioctl_ns(2)](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).

Applications can control the extent to which a `lxkns` discovery tries to
ferret out namespaces from the nooks and crannies of Linux hosts.

> [!NOTE] Some discovery methods are more expensive than others, especially the
> discovery of bind-mounted namespaces in other mount namespaces. The reason
> lies in the design of the Go runtime which runs multiple threads and Linux not
> allowing multi-threaded processes to switch mount namespaces. In order to work
> around this constraint, `lxkns` must fork and immediately re-execute the
> process it is used in just to make it sleep (there's an optional separate
> minimized [`mntnssandbox`](mntnssandbox) binary for this to further reduce
> system resource consumption).

## Required Capabilities

**lxkns** discoveries require the following capabilities:

- `CAP_SYS_PTRACE` grants access to the namespace information in the `/proc`
  file system, as well as access to the file system in other mount namespaces.

- `CAP_SYS_CHROOT` and `CAP_SYS_ADMIN` grant switching into other (especially
  mount) namespaces in order to look into more places compared to standard
  discovery tools.
  
  Additionally, these capabilities allows a discovery service to switch back
  into the initial cgroup namespace in order to discover correct cgroup
  hierarchy information.
  
  Similarly, temporarily switching into the initial mount namespace allows us to
  correctly pick up the freezer ("fridge") states of processes, this works
  around having to either explicitly mount the host's cgroup into the container
  or to unprotect the container's system paths (which docker-compose yet does
  not support).

- `CAP_DAC_READ_SEARCH` grants discovering bind-mounted namespaces without
  interference by any DAC, or "(in)descretionary axcess control".

- `CAP_OVERRIDE` allows access to container engine APIs even as the discovery
  service runs as non-root (and non-Docker user, et cetera).

#### Notes

[^1]: `lsns --task $TID` doesn't seem to work at all (as per lsns from
    util-linux 2.37.2).