/*
Package industrialedge decorates the composer-project flavor of [Siemens
Industrial Edge] apps (“IE apps”) when an IE App project is detected.
Additionally, this decorator also decorates the “Industrial Edge runtime”
container.

This Decorator must be run only after the (Docker) composer project decorator
([github.com/thediveo/lxkns/decorator/composer]), as it relies on the composer
groups having already been detected and their [model.Group] objects created.

# Group Flavor

  - The flavor of composer project groups for Industrial Edge apps (also) is
    "com.siemens.industrialedge.app". Please note that the group type is (still)
    "com.docker.composer.project".

# Container Flavor

  - The flavor of the container housing the Industrial Edge runtime is
    "com.siemens.industrialedge.runtime".
  - The flavor of containers belonging to Industrial Edge apps is
    "com.siemens.industrialedge.app".

[Siemens Industrial Edge]: https://siemens.com/industrial-edge
*/
package industrialedge
