# Linux-kernel Namespaces (+Containers)

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/lxkns)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)
![build and test](https://github.com/thediveo/lxkns/workflows/build%20and%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

![lxkns logo](_images/lxkns-gophers.jpeg ':size=150')

**lxkns** discovers Linux-kernel namespaces as well as mount points in mount
namespaces and relates them to containers, where possible. In every nook and
cranny of your Linux hosts.

Discovery can be operated as a stand-alone REST service or integrated into
system diagnosis tools that need an unobstructed view on Linux-kernel
namespaces.

When it comes to mount namespaces, **lxkns** finds mount points even in
process-less
mount namespaces (for instance, as utilized in ["snap"
technology](https://snapcraft.io/docs)). Our discovery engine even determines
the visibility of mount points, taking different forms of "overmounting" into
consideration.

In a nutshell, **lxkns** is:

- a Go module for discovering namespaces and mount points, with the ability to
  relate namespaces to containers,
- a Go package for switching namespaces (including safely returning),
- a REST API service return JSON discovery results, with an additional web-based
  user interface,
- a set of CLI tools, such as our `lsns`-on-drugs example in `examples/lsallns`
  that lists *all* namespaces with their corresponding containers.

  ```console
  NAMESPACE  TYPE   CONTAINER     PID     PROCESS         COMMENT
  4026531835 cgroup               1       systemd         /init.scope
  4026532338 mnt    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532339 uts    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532340 ipc    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532342 net    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  ```

And now for the obligatory eye candy, click or tap to enlarge:

![lxkns teaser](_images/teaser.png ':class=teaser')
![mount points teaser](_images/teaser-mountpoints.png ':class=teaser')

> [!NOTE] The name **lxkns** derives from **L**inu**x** **k**ernel
> **n**ame**sp**aces. Simply naming it "namespaces" instead would have been a
> too generic name.
