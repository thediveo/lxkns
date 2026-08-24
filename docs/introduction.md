# Linux-kernel Namespaces

..._and_ containers _and_ mount point hierarchies _and_ process CPU-affinities
_and_...

[![Manual](https://img.shields.io/badge/view-manual-blue)](https://thediveo.github.io/lxkns)
[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/lxkns)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)

![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)
![file descriptors](https://img.shields.io/badge/file%20descriptors-not%20leaking-success)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

![lxkns logo](_images/lxkns-gophers.png ':size=150')

Curious...?

- ...do you want to see how Containers use Linux (kernel) namespaces?
- ...learn how these namespaces are used throughout the Linux host, outside of
  containers?
- ...finding out how containers separate mount points, yet how some of these
  mount points nevertheless propagate between host and containers?
- ...seeing which processes and (kernel) tasks are allowed to execute on all or
  only certain CPUs of your system? 

## Abstract

**lxkns**[^1] discovers Linux-kernel namespaces and how they are used by
processes and threads.  In (almost) every nook and cranny of your Linux
hosts[^2]. It then relates the namespaces, processes and threads to containers,
where possible. **lxkns** additionally discovers mount points in mount
namespaces. 

The discovery can be operated as a stand-alone REST service with additional web
UI. Or it can be integrated into system diagnosis tools that need an
unobstructed view on Linux-kernel namespaces.

When it comes to mount namespaces, **lxkns** finds mount points even in
process-less mount namespaces – for instance, as utilized in ["snap"
technology](https://snapcraft.io/docs). This discovery engine even determines
the **visibility of mount points**, taking different forms of **overmounting**
into consideration.

And since **lxkns** needs to discover all processes and tasks anyway, it can
additionally show you which of them activate realtime scheduling and high
priorities, as well the CPUs they're allowed to run on. 

## Quick Deploy

### Docker Compose v2.34.0+ (March 2025)

Make sure that you have a fairly recent Docker engine installed, including the
Docker compose v2 plugin or later. As usual, _Debian_ users are advised to
install docker-ce packages instead of Debian's own _constantly outdated_
docker.io ones. The following single command automatically pulls the compose
deployment as well as the needed image layers and then deploys the default
setup:

```bash
docker compose -f oci://github.com/thediveo/lxkns/app up -d
```

Now point your web browser to `http://localhost:5010` after the service has been
successfully deployed.

![lxkns teaser](_images/teaser-all-namespaces.png ':class=teaser')
![mount points teaser](_images/teaser-mountpoints.png ':class=teaser')

(_Please click or tap the teaser images to enlarge them._)

### Docker Compose v5.2.0+ (June 2026)

**Rejoice!** Starting with Docker compose v5.2.0 or later you can now customize
the IP address the `lxkns` service is bound to as well as its HTTP port. Simply
pass one or both environment variables to `docker compose ... up -y`:

| Env var | Default |
| --- | --- |
| `LXKNS_IP` | `127.0.0.1` |
| `LXKNS_PORT` | `5010` |

For instance, to deploy to a different port on `127.0.0.1:5050` (`-y`
automatically accepts all configured settings):

```bash
LXKNS_PORT=5050 docker compose -f oci://ghcr.io/thediveo/lxkns/app/cfg:latest up -y
```

> [!NOTE] The customizable OCI artifact is `ghcr.io/thediveo/lxkns/app/cfg`, not
> `ghcr.io/thediveo/lxkns/app`. A separate OCI artifact is necessary so that
> older compose version still work with the non-customizable OCI artifact, as
> they get tripped up by the customizable artifact.

## Technical

**lxkns** is...

- a Go module for discovering **namespaces** and **mount points**, with the
  ability to relate namespaces to **containers**,
- a Go package for **switching namespaces** (including safely returning),
- another Go package for **directly(!) reading from other mount namespaces**
  using ordinary file operations,
- a **REST API discovery service** returns JSON discovery results, with an
  additional **web-based user interface** as the icing on top,
- a set of **CLI tools**, as well as our `lsns`-on-drugs example in
  `examples/lsallns` that lists *all* namespaces with their corresponding
  *containers*.
  ```console
NAMESPACE  TYPE   CONTAINER     PID   PROCESS/[TASK]    COMMENT
4026531835 cgroup               1     systemd           cgroup:/init.scope
4026532740 mnt    lxkns-lxkns-1 12706 lxkns             cgroup:/system.slice/docker-88c84e7f9e668b6ea63af1e9b496b265638f5b0493bc20c3e406e438f53a9e05.scope
4026532741 uts    lxkns-lxkns-1 12706 lxkns             cgroup:/system.slice/docker-88c84e7f9e668b6ea63af1e9b496b265638f5b0493bc20c3e406e438f53a9e05.scope
4026532742 ipc    lxkns-lxkns-1 12706 lxkns             cgroup:/system.slice/docker-88c84e7f9e668b6ea63af1e9b496b265638f5b0493bc20c3e406e438f53a9e05.scope
4026532743 net    lxkns-lxkns-1 12706 lxkns             cgroup:/system.slice/docker-88c84e7f9e668b6ea63af1e9b496b265638f5b0493bc20c3e406e438f53a9e05.scope
4026532861 net                  9313  [stray thread :p] cgroup:/user.slice/user-1000.slice/user@1000.service/app.slice/snap.code.code.b9d9c611-0444-4f49-9566-ea28bef2e6a4.scope
  ```

#### Notes

[^1]: The name **lxkns** derives from **L**inu**x** **k**ernel
      **n**ame**sp**aces. Simply naming it "namespaces" instead would have been
      a too generic name. And Go is *very* opinionated when it comes to module
      names that are "too long", "too generic", or "not generic" enough. Alice must
      have gone down a Gopher hole, Lewis Caroll didn't got that part right.

[^2]: Michael Kerrisk of "The Linux Programming Interface" and
    [man7.org](https://man7.org) fame nudged me on my claim enough so that
    **lxkns** as of v0.24 finally got the missing task discovery support.
