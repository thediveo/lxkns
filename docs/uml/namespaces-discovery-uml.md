```plantuml
hide empty fields
hide empty methods

namespace lxkns {
  
  class Discover as " " <<(F,LightGray)>> {
    Discover(opts DiscoverOpts) *DiscoveryResult
  }
  
  Discover <.. DiscoverOpts : "controls"
  Discover ..> DiscoveryResult : "returns"
  
  class "DiscoverOpts" <<(S,YellowGreen)>> {
    NamespaceTypes species.NamespaceType
    SkipProcs bool
    SkipTasks bool
    SkipFds bool
    SkipBindmounts bool
    SkipHierarchy bool
    SkipOwnership bool
  }
  
  class "DiscoveryResult" <<(S,YellowGreen)>> {
    Options DiscoverOpts
    Namespaces AllNamespaces
    InitialNamespaces NamespacesSet
    ' TODO: Root(s)
    Processes ProcessTable
  }
  
}
```
