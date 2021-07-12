/*

Package kuhbernetes provides Decorators for "recovering" Kubernetes pods from
the containers found.

Pod Group Decoration

A Group representing the containers of a pod is typed as "io.kubernetes.pod".

Container Decoration

Containers that are part of a Kubernetes pod are decorated with the pod's
namespace, pod's name, and individual container name:

    * "io.kubernetes.pod.namespace"
    * "io.kubernetes.pod.name"
    * "io.kubernetes.container.name"

Especially in the case of dockershim-managed containers, the
"io.kubernetes.container.name" specifies the user-defined container name as
opposed to the Docker container name, where the latter is under sole control of
the dockershim.

Sandbox containers (also termed "pause" containers) are containers created at
least in some container engine environments in order to correctly set up a pod's
networking before starting any user containers. These infrastructure containers
are labelled with:

    * "lxkns/k8s/container/kind"

Please note that this label is supported by different decorators and thus not
limited to containerd CRI-based system configurations, but also applies to, for
instance, dockershim-based system configurations.

Trivia

Isn't there a typo in this package name? Actually not. Kubernetes is all about
kettle, not cats. And cows (ger. "Kühe") are kettle.

*/
package kuhbernetes
