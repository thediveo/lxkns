# Deployment

Features of deploying the containerized lxkns service:

- **read-only:** the lxkns service can be used on a read-only container file
  system without any issues.

- **non-root:** the holy grail of container hardening … wait till you get to see
  our capabilities below. Ah, the sweet scent of security snake oil.

- **unprivileged:** because that doesn't mean in-capable when not using the
  `--privileged` nuke option.

- **capabilities:** not much to see here, just…
  - `CAP_SYS_PTRACE` gives us access to the namespace information in the proc
    file system.
  - `CAP_SYS_CHROOT` and `CAP_SYS_ADMIN` allow us to switch (especially mount)
    namespaces in order to look into more places compared to standard discovery
    tools.
    
    Additionally, they allow us to switch the discovery service back into the
    initial cgroup namespace in order to discover correct cgroup hierarchy
    information.
    
    Similar, temporarily switching into the initial mount namespace allows us to
    correctly pick up the freezer ("fridge") states of processes, this works
    around having to either explicitly mount the host's cgroup into the
    container or to unprotect the container's system paths (which docker-compose
    yet does not support).
  - `CAP_DAC_READ_SEARCH` allows us to discover bind-mounted namespaces without
    interference by any in-descretionary excess control (DAC).
  - `CAP_DAC_OVERRIDE` allows us to connect to the containerd API socket without
    being root.

The containerized **lxkns** service correctly handles these pitfalls:

- **reading from other mount namespaces**: in order to discover mount points
  from a process-less bind-mounted mount namespace, lxkns usually simply kicks
  off a new thread that then attaches itself to the bind-mounted mount namespace
  and does nothing more. With this "sandbox" thread idling, lxkns then can read
  from the bind-mounted mount namespace (such as its `mountinfo`) via the Linux
  kernel's procfs "wormholes": `/proc/[PID]/root/...`, see also
  [proc(5)](https://man7.org/linux/man-pages/man5/proc.5.html). In case a mount
  namespace is in a different user namespace, lxkns uses a separate "sandbox"
  process instead – which it can create by simply forking and re-executing
  itself.

- **cgroup namespaced container**: during startup, lxkns detects when it has
  been placed into its own cgroup namespace ... as, for example, it is the case
  in newer Docker default installations on Linux base OS configurations
  especially with a cgroups v2 unified hierarchy. Without further measures, lxkns would be unable to discover the correct freezer states of processes. Thus, lxkns then switches itself out of its own cgroup namespace and back into the host's initial namespace, if possible. Please note that running lxkns in a non-initial namespace blocks correct discovery, not least process freezer state discovery.
