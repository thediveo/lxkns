/*
Package containerizer provides the implementations to store data about
[Container] and [ContainerEngine] objects. Subpackages provide so-called
“containerizers” (implementing [Containerizer]) that discover Containers from
specific container engines.

Nota bene: the lxkns model defines only the [model.Container] and
[model.ContainerEngine] interfaces instead of implementations in order to break
import cycles that otherwise would invariably occur.

Nota bene-bene: sometimes, preemptive interfaces are simply necessary, because
they don't have any idea of Golotry. especially, when the domain-specific
(namespace) language uses type inheritance and embedding doesn't eat the cake.
*/
package containerizer
