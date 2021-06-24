```plantuml
hide empty fields
hide empty methods

class "Flat Linux Kernel Namespace" as ns <<(L,LightBlue)>>

class cgroup <<(L,LightBlue)>>
note bottom: control group
ns <|--- cgroup
class ipc <<(L,LightBlue)>>
note bottom: SYSV\ninter-process\ncommunication
ns <|--- ipc
class mnt <<(L,LightBlue)>>
note bottom: filesystem\nmounts
ns <|--- mnt
class net <<(L,LightBlue)>>
note bottom: network
ns <|--- net
class uts <<(L,LightBlue)>>
note bottom: hostname and\ndomain name
ns <|--- uts
class time <<(L,LightBlue)>>
note bottom: monotonic +\nboot-time clocks
ns <|--- time

cgroup -[hidden] ipc

class "Hierarchical Namespace" as hns <<(L,LightBlue)>>
ns <|-- hns

class pid <<(L,LightBlue)>>
note bottom: PIDs
hns <|-- pid
hns <--> hns : parent/children

class user <<(L,LightBlue)>>
note bottom: uids/gids,\ncaps, …
hns <|-- user
ns <-- pid : owns

user -[hidden] pid
```
