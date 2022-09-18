/*
Package dockershim decorates Kubernetes pod groups discovered from Docker
container names managed by the (in)famous Docker shim.

# The Kubernetes Dockershim

The so-called “dockershim” uses especially crafted Docker container names to
encode pod-related information without the need for an additional database. This
“stateless” (or, database-less) design allows us to regenerate some Kubernetes
pod information (name, namespace, container name from the k8s perspective) given
just container names.

# Background Information

Docker doesn't seem to have any hard restrictions as to the length of container
names. However, it restricts the allowed characters in container names. Simply
spoken, Docker container names can consist of lower and upper case aA-zZ, digits
0-9, and finally dashes, underscores and dots. Please note these last three
characters cannot be in the first position though. See also [Docker:
restrictions regarding naming container].

The Kubernetes dockershim encodes [pod-related information in Docker container
names] as follows:

	k8s_<containername>_<metadata.name>_<metadata.namespace>_<metadata.uid>_<attempt>[_<random>]

Kubernetes restricts the pod name, namespace, and container name to consist only
of lower case a-z, but does not allow uppercase A-Z. It additionally restricts
them to the maximum length of DNS labels, that is, 63 characters (not: glyphs).

The special “pause” (sandbox) pod gets the reserved "POD" name. Since Kubernetes
only allow lower case letters in container names, this ensures that there never
can be a conflicting user container also named "POD", only a non-conflicting
"pod". See also the aptly named [leaky.go] definition.

As the metadata.uid field can use different uid schemes, don't rely on a
specific format. Just take it as a Docker-conforming string, nothing more. It
cannot contain underscores, as these are already used for separating the
individual pod data fields.

The attempt field is of no interest to us, as it is related to the so-called
sandbox (=pause container) management.

The random appendix only appears in case of Docker somehow loosing its mind due
to the Docker container name conflict bug. It seems to be present in Docker
versions up to 1.11, at least the Kubernetes Docker shim seems to suggest this.
See see also the [details of the (closed) Docker bug].

[details of the (closed) Docker bug]: https://github.com/moby/moby/issues/23371
[Docker: restrictions regarding naming container]: https://stackoverflow.com/questions/42642561/docker-restrictions-regarding-naming-container
[pod-related information in Docker container names]: https://github.com/kubernetes/kubernetes/blob/7f23a743e8c23ac6489340bbb34fa6f1d392db9d/pkg/kubelet/dockershim/naming.go#L29
[leaky.go]: https://github.com/kubernetes/kubernetes/blob/2e357e39c81673f916a81a0a4f485ed080043e25/pkg/kubelet/leaky/leaky.go
*/
package dockershim
