# Deployment

Features of deploying the containerized lxkns service:

- **read-only:** the lxkns service can be used on a read-only container file
  system without any issues.

- **non-root:** the holy grail of container hardening … wait till you get to
  see our capabilities below.

- **unprivileged:** because that doesn't mean in-capable.

- **capabilities:** not much to see here, just…
  - `CAP_SYS_PTRACE` gives us access to the namespace information in the proc
    file system.
  - `CAP_SYS_ADMIN` and `CAP_SYS_ADMIN` allow us to switch (especially mount)
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

The convertainerized lxkns service correctly handles these pitfalls:

- **reading from other mount namespaces**: in order to discover mount points
  from a process-less bind-mounted mount namespace, lxkns forks itself and then
  re-executes the child in the mount namespace to read its procfs `mountinfo`
  from. The child here acts as the required procfs entry to be able to read the
  correct `mountinfo` at all. However, when containerized, lxkns runs in its own
  mount namespace, whereas the bindmount of the mount namespace will be in some
  other mount namespace, such as the host's initial mount namespace. In order to
  successfully reference the bindmount in the VFS, lxkns uses the Linux kernel's
  procfs wormholes: `/proc/[PID]/root/...`, see also
  [proc(5)](https://man7.org/linux/man-pages/man5/proc.5.html).

- **cgroup namespaced container**: during startup, lxkns detects when it has
  been placed into its own cgroup namespace ... as, for example, it is the case
  in newer Docker default installations on Linux base OS configurations
  especially with a cgroups v2 unified hierarchy. Without further measures, lxkns would be unable to discover the correct freezer states of processes. Thus, lxkns then switches itself out of its own cgroup namespace and back into the host's initial namespace, if possible. Please note that running lxkns in a non-initial namespace blocks correct discovery, not least process freezer state discovery.
