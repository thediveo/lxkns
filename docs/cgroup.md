# Cgroup Discovery

In addition to namespaces and their related processes, lxkns also discovers the
freezer state and (freezer) cgroup controller path information for the processes
attached to namespaces.

> [!ATTENTION] When the lxkns service is deployed containerized using the included
> `deployments/lxkns/docker-compose.yaml` it will automatically detect system
> configurations where **Docker places the service container in its own cgroup
> namespace**. In order to provide full discovery information **the lxkns
> service then automatically switches back out of the container's cgroup
> namespace into the initial cgroup namespace**. The rationale here is that
> docker-compose unfortunately lacks support for the `--cgroupns=host` CLI flag
> ([issue #8167](https://github.com/docker/compose/issues/8167)) and thus
> switching is necessary as a stop-gap measure.

## v1, v1+v2, v2

The lxkns discovery engine transparently handles the gory differences between
cgroups v1 and v2, even in hybrid configurations. As a rule of thumb, the main
basic difference is that in cgroups v1 the freezers are a dedicated controller
type, while in cgroups v2 the freezers are now an integrated core controller
functionality.

- **cgroups v1**: lxkns automatically detects the v1 freezer hierarchy and its
  location.

- "hybrid" **cgroups v1+v2**: if the v1 freezer hierarchy is mounted, lxkns will
  automatically detect and use it; otherwise, lxkns will use the unified cgroups
  v2 hierarchy instead, with its built-in core freezer (Linux kernel 5.2 and
  later).

- "pure" **cgroups v2**: lxkns will automatically detect and use the unified
  cgroups v2 hierarchy when no v1 freezer hierarchy is present.

## Control Group Paths

To a limited extend, the paths of (freezer) control groups often relate to the
partitioning of processes using Linux kernel namespaces. For instance, processes
in Docker containers will show control group names in the form of `docker/<id>`
(or `docker-<id>.scope`, or ...), where the id is the usual 64 hex char string.
Plain containerd container processes will show up with `<namespace>/<id>`
control group names.
