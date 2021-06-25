# Linux-kernel Namespaces

[![PkgGoDev](https://pkg.go.dev/badge/github.com/thediveo/lxkns)](https://pkg.go.dev/github.com/thediveo/lxkns)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)
![build and test](https://github.com/thediveo/lxkns/workflows/build%20and%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

![lxkns logo](_images/lxkns-gophers.jpeg ':size=150')

**lxkns** discovers Linux-kernel namespaces as well as mount points in mount
namespaces. In every nook and cranny of your Linux hosts.

For mount namespaces, lxkns finds mount points even in process-less mount
namespaces (for instance, as utilized in ["snap"
technology](https://snapcraft.io/docs)). Our discovery engine even determines
the visibility of mount points, taking different forms of "overmounting" into
consideration.

In a nutshell, lxkns is:

- a Golang module for discovering and switching namespaces,
- a REST API service with an additional web-based user interface,
- a set of CLI tools.

> [!NOTE] The name **lxkns** derives from **L**inu**x** **k**ernel
> **n**ame**sp**aces.
