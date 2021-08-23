# Discovery Examples

These are simple examples to hopefully give you a good first impression. For the
real-world gory stuff, please take a look at the `examples/` and `cmd/`
directories in the lxkns repository. 😇

## Namespaces Only

This simple example code runs a full namespace (-only) discovery and then prints
all namespaces found, sorted by their type, then by their ID. It ignores
containers completely, being the "lie in user space" they are.

```go
package main

import (
    "fmt"
    "github.com/thediveo/lxkns/discover"
    "github.com/thediveo/lxkns/model"
)

func main() {
    result := discover.Discover(discover.StandardDiscovery())
    for nsidx := model.MountNS; nsidx < model.NamespaceTypesCount; nsidx++ {
        for _, ns := range result.SortedNamespaces(nsidx) {
            fmt.Println(ns.String())
        }
    }
}
```

The lxkns module hides lots of really ugly and truely gory discovery process
details, so API users can focus on making good use of the information discovered
instead of getting lost in the low-level craziness of namespace hunting and
switching.

## With Containers

This simple example code (from `examples/barebones`) runs a full namespace
discovery including "containerization" and then prints all namespaces found,
sorted by their type, then by their ID. When a namespace is associated with a
container, then the container's name will also be printed.

```go
package main

import (
    "context"
    "fmt"

    "github.com/thediveo/lxkns/containerizer/whalefriend"
    "github.com/thediveo/lxkns/discover"
    "github.com/thediveo/lxkns/model"
    "github.com/thediveo/whalewatcher/watcher"
    "github.com/thediveo/whalewatcher/watcher/moby"
)

func main() {
    // Set up a Docker engine-connected containerizer and wait for it to
    // synchronize.
    moby, err := moby.New("", nil)
    if err != nil {
        panic(err)
    }
    <-moby.Ready()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    cizer := whalefriend.New(ctx, []watcher.Watcher{moby})

    // Run the discovery, including containerization.
    result := discover.Discover(
        discover.WithStandardDiscovery(), discover.WithContainerizer(cizer))

    for nsidx := model.MountNS; nsidx < model.NamespaceTypesCount; nsidx++ {
        for _, ns := range result.SortedNamespaces(nsidx) {
            fmt.Println(ns.String())
        }
    }
}
```

Please note that in this case it is necessary to explicitly wait for the
container engine adapter (`moby.New()`) to become synchronized, as otherwise
discovery results might yield spurious results depending on system load. This
wait might be skipped in a service (such as lxkns), where the discovery API is
designed as "best effort" in order to get the service serving even if not all
container engines are yet online or will never be (depending on system
configuration).
