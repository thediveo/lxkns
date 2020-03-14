```plantuml
hide empty fields
hide empty methods
!define L <size:12><&link-intact></size><i>

interface Hierarchy {
  L Parent() Hierarchy
  L Children() []Hierarchy
}

interface Ownership {
  UID() int
  L Ownings() AllNamespaces
}

Hierarchy "*" -up-> Hierarchy : Parent
Hierarchy <-down- "*" Hierarchy : Children

Hierarchy -[hidden] Ownership

Ownership --> "*" Namespace : "Ownings"

interface Namespace {
  ID() nstypes.NamespaceID
  Type() nstypes.NamespaceType
  L Owner() Hierarchy
  Ref() string
  L Leaders() []*Process
  LeaderPIDs() []PIDType
  L Ealdorman() *Process
  String() string
}

Hierarchy <-- Namespace : "Owner"

interface NamespaceStringer {
  TypeIDString() string
}
Namespace <|- NamespaceStringer
```
