# (Un)Marshalling Example

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
    b, _ := json.Marshal(
        apitypes.NewDiscoveryResult(
            lxkns.Discover(lxkns.StandardDiscovery())))

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
