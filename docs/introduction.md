# Linux-kernel Namespaces (+Containers)

[![Manual](https://img.shields.io/badge/view-manual-blue)](https://thediveo.github.io/lxkns)
[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/lxkns)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)

![build and test](https://github.com/thediveo/lxkns/workflows/build%20and%20test/badge.svg?branch=master)
![Coverage](https://img.shields.io/badge/Coverage-87.7%25-brightgreen)
![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)
![file descriptors](https://img.shields.io/badge/file%20descriptors-not%20leaking-success)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

![lxkns logo](_images/lxkns-gophers.jpeg ':size=150')

## Abstract

**lxkns**[^1] discovers Linux-kernel namespaces as well as mount points in mount
namespaces. It then relates them to containers, where possible. In (almost)
every nook and cranny of your Linux hosts.

This discovery can be operated as a stand-alone REST service with additional web
UI. Or it can be integrated into system diagnosis tools that need an
unobstructed view on Linux-kernel namespaces.

When it comes to mount namespaces, **lxkns** finds mount points even in
process-less mount namespaces – for instance, as utilized in ["snap"
technology](https://snapcraft.io/docs). Our discovery engine even determines the
**visibility of mount points**, taking different forms of **overmounting** into
consideration.

> [!WARNING] Please check [Important Changes](changelog), especially if you
> have been used the API in the past, and not only the service.

## Eye Candy

The lxkns service provides a web user interface for comfy discovery.

![lxkns teaser](_images/teaser.png ':class=teaser')
![mount points teaser](_images/teaser-mountpoints.png ':class=teaser')

(Please click or tap to enlarge)

## In a Nutshell

**lxkns** is...

- a Go module for discovering **namespaces** and **mount points**, with the
  ability to relate namespaces to **containers**,
- a Go package for **switching namespaces** (including safely returning),
- a **REST API discovery service** return JSON discovery results with an
  additional **web-based user interface** on top,
- a set of **CLI tools**, such as our `lsns`-on-drugs example in
  `examples/lsallns` that lists *all* namespaces with their corresponding
  *containers*.
  ```console
  NAMESPACE  TYPE   CONTAINER     PID     PROCESS         COMMENT
  4026531835 cgroup               1       systemd         /init.scope
  4026532338 mnt    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532339 uts    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532340 ipc    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  4026532342 net    lxkns_lxkns_1 1452785 lxkns           /docker/58ba4492582ab4a938646ec3dd2328e3866d42f9ac47bcc9f9693fc4c2479047
  ```

#### Notes

[^1]: The name **lxkns** derives from **L**inu**x** **k**ernel
      **n**ame**sp**aces. Simply naming it "namespaces" instead would have been
      a too generic name. And Go is *very* opinionated when it comes to module
      names that are too long, too generic, or not generic enough. Alice must
      have gone down a Gopher hole, Lewis Caroll didn't got that part right.
