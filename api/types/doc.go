/*
Package types defines the common types for (un)marshalling elements of the lxkns
information model from/to JSON.

  - PIDMap wraps model.PIDMap
  - ProcessTable wraps model.ProcessTable
  - (Process wraps model.Process, but is not intended for direct consumption)

# Discovery Results

Most lxkns API users probably want to simply marshal and unmarshal discovery
results without any hassle. So, here we go:

To marshall a given [discover.Result] in a service:

	allns := discover.Namespaces(discover.StandardDiscovery())
	err := json.Marshal(NewDiscoveryResult(WithResult(allns)))

And then to unmarshall a discovery result into “allns” when consuming a
discovery service:

	disco := NewDiscoveryResult()
	err := json.Unmarshal(jsondata, disco)
	allns := disco.Result()

# Process Table

Process Tables of type [model.ProcessTable] are un/marshalled from/to JSON with
the help of the [ProcessTable] type from this package.

Because the Linux architecture closely couples processes and namespaces, all
processes in the process table always reference namespaces. This poses a slight
difficulty when unmarshalling, because we need to deal with two separate
elements which can unmarshalled only sequentially, not simultaneously (whatever
“simultaneously” would actually mean). Because we have no control over the
order in which the process table and the namespaces will be unmarshalled,
whenever we get references to namespaces which we cannot resolve yet, we simply
“pre-create” namespace objects (we prime the namespace map). These will then
later be completely unmarshalled. For this to work, a [ProcessTable] needs a
reference to a [NamespacesDict] in order to pre-create the correct namespace
objects (ID and type only).

Fortunately, it is also possible to un/marshal a stand-alone [ProcessTable]
only; in this case the unmarshalled process table will reference only the
pre-created minimal Namespace objects. Such minimal Namespace objects contain
only a valid ID and type; all other information is missing (zero values).

# PID Maps for Translating PIDs between PID Namespaces

PID maps of type [model.PIDMap] are un/marshalled from/to JSON with the help of
the PIDMap type from this package.

To marshal a [model.PIDMap], create a wrapper object (“Digital Twin”) and
marshal the wrapper as you need:

	// This is one way to get a PID map to be marshalled next.
	pidmap := model.NewPIDMap(discover.Discovery(discover.FullDiscovery))

	// Wrap the PID map and then marshal it...
	out, err := json.Marshal(NewPIDMap(WithPIDMap(allpidmap)))

Unmarshalling can be done either without or with an additional PID namespace
context. Without a PID namespace context is useful when just unmarshalling the
PID map and the PID namespaces need only to be known in terms of their IDs (and
PID type), but without further namespace details (which isn't included in the
PID map anyway).

	pm := NewPIDMap()
	err := json.Unmarshal(out, &pm)
	pidmap := pm.PIDMap // access wrapped model.PIDMap
*/
package types
