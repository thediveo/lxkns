# API

These are simple examples to give you a first impression. For the real-world
gory stuff, please take a look at the `examples/` and `cmd/` directories in the
lxkns repository. 😇

## 🔎 Discovery

This simple example code runs a full namespace discovery and then prints all
namespaces found, sorted by their type, then by their ID.

```go
package main

import (
    "fmt"
    "github.com/thediveo/gons/reexec"
    "github.com/thediveo/lxkns"
    "github.com/thediveo/lxkns/model"
)

func main() {
    reexec.CheckAction() // must be called before a full discovery
    result := lxkns.Discover(lxkns.FullDiscovery)
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

## 📡 Marshalling and Unmarshalling

`lxkns` supports un/marshalling discovery results from/to JSON. Both the
namespaces and process information can be passed via JSON and correctly
regenerated.

```go
package main

import (
    "fmt"
    "github.com/thediveo/gons/reexec"
    "github.com/thediveo/lxkns"
    apitypes "github.com/thediveo/lxkns/api/types"
)

func main() {
    reexec.CheckAction() // only for discovery, not for unmarshalling
    b, _ := json.Marshal(apitypes.NewDiscoveryResult(lxkns.Discover(lxkns.FullDiscovery)))

    dr := apitypes.NewDiscoveryResult(nil)
    _ = json.Unmarshal(b, &dr)
    result := (*lxkns.DiscoveryResult)(dr)
}
```

> [!NOTE] Discovery results need to be explicitly "wrapped" in JSON-able objects
> for un/marshalling. The discovery result objects returned from
> `lxkns.Discover()` cannot be properly un/marshalled, not least as they
> describe an information model with circular references that is optimized for
> quick navigation, not for un/marshalling.
