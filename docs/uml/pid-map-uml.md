```plantuml
hide empty fields
hide empty methods
!define L <size:12><&link-intact></size><i>

package "lxkns" {

class "PIDMap" <<(S,YellowGreen)>> {
  m map[NamespacedPID]NamespacedPIDs
  Translate(pid PIDType, from Namespace, to Namespace) PIDType
}

PIDMap -[hidden]- NamespacedPID
PIDMap -[hidden]-- NamespacedPIDs

class NamespacedPIDs <<(T,Khaki)>> {
    []NamespacedPID
}
note left {
    slice of <PID; namespace ID> pairs
    which refer to the same process
    in its own PID namespace, as well as
    in all parent PID namespaces.
}

class NamespacedPID {
  PIDNS Namespace
  PID PIDType
}
note right: a PID and the namespace ID\nthe PID is valid in.

}

object ": PIDMap" as pidmap {
  m
}
NamespacedPIDs -[hidden]- pidmap

object ": NamespacedPIDs" as pid1 {
  [1]: {PID: 1, PIDNS: 1}
}
pidmap -[hidden]- pid1
pidmap -- pid1 : "{PID: 1, PIDNS: 1} >"

object ": NamespacedPIDs" as pid4567 {
  [0]: {PID: 4567, PIDNS: 1}
  [1]: {PID: 1, PIDNS: 2}
}
pidmap -[hidden]- pid4567
pidmap --- pid4567 : "{PID: 4567, PIDNS: 1} >"
pidmap --- pid4567 : "{PID: 1, PIDNS: 2} >"

```
