# Linux kernel Namespaces

[![GoDoc](https://godoc.org/github.com/TheDiveO/lxkns?status.svg)](http://godoc.org/github.com/TheDiveO/lxkns)
[![Architecture](https://img.shields.io/badge/doc-architecture-blue)](docs/architecture.md)
[![GitHub](https://img.shields.io/github/license/thediveo/lxkns)](https://img.shields.io/github/license/thediveo/lxkns)
![build and test](https://github.com/TheDiveO/lxkns/workflows/build%20and%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/lxkns)](https://goreportcard.com/report/github.com/thediveo/lxkns)

`lxkns` is a Golang package for discovering Linux kernel namespaces. In every
nook and cranny of your Linux hosts.

## Comprehensive Namespace Discovery

When compared to most well-known and openly available CLI tools, such as
`lsns`, the `gons` package detects namespaces even in places of a running
Linux system other tools typically do not consider. In particular:

1. from the procfs filesystem in `/proc/[PID]/ns/*` -- as `lsns` and other tools do.
2. bind-mounted namespaces, via `/proc/[PID]/mountinfo`. Our discovery method
   even finds bind-mounted namespaces in _other_ mount namespaces than the
   current one in which the discovery starts.
3. file descriptor-referenced namespaces, via `/proc/[PID]/fd/*`.
4. intermediate hierarchical user and PID namespaces, via `NS_GET_PARENT`
   ([man 2 ioctl_ns](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).
5. user namespaces owning non-user namespaces, via `NS_GET_USERNS` ([man 2
   ioctl_ns](http://man7.org/linux/man-pages/man2/ioctl_ns.2.html)).

| tool | `/proc/[PID]/ns/*` ① | bind mounts ② | `/proc/[PID]/fd/*` ③ | hierarchy ④ | owning user namespaces ⑤ |
| -- | -- | -- | -- | -- | -- |
| `lsns` | ✓ | | | |
| `lxkns` | ✓ | ✓ | ✓ | ✓ | ✓ |

Applications can control the extent to which a `lxkns` discovery tries to
ferret out namespaces from the nooks and crannies of Linux hosts.

> Some discovery methods are more expensive than others, especially the
> discovery of bind-mounted namespaces in other mount namespaces. The reason
> lies in the design of the Go runtime which runs multiple threads and Linux
> not allowing multi-threaded processes to switch mount namespaces. In order
> to work around this constraint, `lxkns` must fork and immediately re-execute
> the process it is used in. Applications that want to use such advanced
> discovery methods thus **must** call `reexec.CheckAction()` as early as
> possible in their `main()` function. For this, you need to `import
> "github.com/thediveo/gons/reexec"`.

## gons CLI tools

But `lxkns` is more than "just" a Golang package. It also features CLI tools
build on top of `lxkns` (we _do_ eat our own dog food):

- `lsuns`: shows _all_ user namespaces in your Linux host, in a neat
  hierarchy. Moreover, it can also show the non-user namespaces "owned" by
  user namespaces. This ownership information is important with respect to
  capabilities and processes switching namespaces using `setns()` ([man 2
  setns](http://man7.org/linux/man-pages/man2/setns.2.html)).

- `lspns`: shows _all_ PID namespaces in your Linux host, in a neat hierarchy.

## Package Usage

The following example code runs a full namespace discovery using
`Discover(FullDiscovery)` and then prints all namespaces found, sorted by
their type, then by their ID.

```go
package main

import (
    "fmt"
    "github.com/thediveo/gons/reexec"
    "github.com/thediveo/lxkns"
)

func main() {
    reexec.CheckAction() // must be called before a full discovery
    result := lxkns.Discover(lxkns.FullDiscovery)
    for nsidx := lxkns.MountNS; nsidx < lxkns.NamespaceTypesCount; nsidx++ {
        for _, ns := range result.SortedNamespaces(nsidx) {
            fmt.Println(ns.String())
        }
    }
}
```

## Copyright and License

`lxkns` is Copyright 2020 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
