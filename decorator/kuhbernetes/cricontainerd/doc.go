/*
Package cricontainerd decorates Kubernetes pod groups discovered from
CRI-managed containers, based on their CRI-related labels.

Please note that at this time only containerd-originating k8s containers are
handled by this decorator. Supporting further CRI runtimes is currently blocked
upstream waiting for a container lifecycle event CRI API to land in k8s.
*/
package cricontainerd
