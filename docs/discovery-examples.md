# Discovery Examples

These are simple examples to give you a first impression. For the real-world
gory stuff, please take a look at the `examples/` and `cmd/` directories in the
lxkns repository. 😇

## Namespaces Only

This simple example code runs a full namespace (-only) discovery and then prints
all namespaces found, sorted by their type, then by their ID.

```go
package main

import (
    "fmt"
    "github.com/thediveo/gons/reexec"
    "github.com/thediveo/lxkns"
    "github.com/thediveo/lxkns/model"
)

func main() {
    reexec.CheckAction() // must be called before a standard discovery
    result := lxkns.Discover(lxkns.StandardDiscovery())
    for nsidx := model.MountNS; nsidx < model.NamespaceTypesCount; nsidx++ {
        for _, ns := range result.SortedNamespaces(nsidx) {
            fmt.Println(ns.String())
        }
    }
}
```

The lxkns module hides lots of ugly and gory discovery process details, so API
users can focus on making good use of the information discovered instead of
getting lost in the low-level craziness of namespace hunting and namespace
switching.

For instance, as the Go runtime is (OS level) multi-threaded, some types of
namespaces cannot be switched during a full discovery (in particular, mount
namespaces). Thus the discovery process internally forks and then immediately
re-executes its own binary: when the re-executed child starts it detects the
restart and switches into another (mount) namespace, just before the Go runtime
spins up, and then carries out a further step of the discovery process. All
these gory details are hidden by the `github.com/thediveo/gons/reexec` package
and its `reeexec.CheckAction()`.

## Containers

This simple example code (from `examples/barebones`) runs a full namespace
discovery including "containerization" and then prints all namespaces found,
sorted by their type, then by their ID. When a namespace is associated with a
container, then the container's name will also be printed.

```go
package main

import (
    "context"
    "fmt"

    "github.com/thediveo/gons/reexec"
    "github.com/thediveo/lxkns"
    "github.com/thediveo/lxkns/containerizer/whalefriend"
    "github.com/thediveo/lxkns/model"
    "github.com/thediveo/whalewatcher/watcher"
    "github.com/thediveo/whalewatcher/watcher/moby"
)

func main() {
    reexec.CheckAction() // must be called before a standard discovery

    // Set up a Docker engine-connected containerizer
    moby, err := moby.NewWatcher("")
    if err != nil {
        panic(err)
    }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    cizer := whalefriend.New(ctx, []watcher.Watcher{moby})

    // Run the discovery, including containerization.
    result := lxkns.Discover(
        lxkns.WithStandardDiscovery(), lxkns.WithContainerizer(cizer))

    for nsidx := model.MountNS; nsidx < model.NamespaceTypesCount; nsidx++ {
        for _, ns := range result.SortedNamespaces(nsidx) {
            fmt.Println(ns.String())
        }
    }
}
```
