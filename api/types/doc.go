/*

Package types defines the common types for (un)marshalling elements of the
lxkns information model from/to JSON.

PID Maps for Translating PIDs between PID Namespaces

PID maps of type lxkns.PIDMap are un/marshalled from/to JSON with the help of
the PIDMap type from this package.

To marshal an lxkns.PIDMap, create a wrapper object ("Digital Twin") and
marshal the wrapper as you need:

    // This is how you might get your PID map...
    pidmap := lxkns.NewPIDMap(lxkns.Discovery(lxkns.FullDiscovery))

    // Wrap your PID map and then marshal it...
    out, err := json.Marshal(NewPIDMap(WithPIDMap(allpidmap)))

Unmarshalling can be done either without or with an additional PID namespace
context. Without a PID namespace context is useful when just unmarshalling the
PID map and the PID namespaces need only to be known in terms of their IDs
(and PID type), but without further namespace details (which isn't included in
the PID map anyway).

    pm := NewPIDMap()
    err := json.Unmarshal(out, &pm)
    pidmap := pm.PIDMap // access wrapped lxkns.PIDMap

*/
package types
