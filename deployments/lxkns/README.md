# lxkns as a Service

Deploys an lxkns-powered namespace discovery service as a container.

- service API is exposed on port `5010` of the host system.
- available service API paths are `/api/{namespaces,processes,pidmap}`.
- read-only container filesystem.
- non-privileged container, running as non-root, yet with some capabilities.
- comes with a dedicated Seccomp profile, derived from Docker's default profile.
- optional AppArmor profile, but as this has to be loaded into the system by a
  system admin, the default operation is to switch off AppArmor for the lxkns
  container.

There's an optional AppArmor profile in file `lxkns-apparmor` available; it must
be loaded into the kernel prior to deploying the containerized lxkns service:

1. `sudo apparmor_parser lxkns-apparmor`
2. edit `docker-compose.yaml` and set the `- apparmor:lxkns` element of
   `security_opt` to activate the previously loaded profile for the lxkns
   container deployment.
