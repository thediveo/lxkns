# All Namespaces View

The default view is named "all namespaces" and shows all discovered namespaces,
organized along the hierarchy of **user namespaces** ➊. The number next to the
title "Linux Namespaces" indicates to total number of all namespaces shown.

![view all namespaces](_images/all-namespaces-view.png ':class=framedscreenshot')

This view reflects the architectural design of the Linux namespaces, where any
namespace is always owned by a user namespace. In case of user namespaces owning
user namespaces this is at the same time also the parent-child relationship.

lxkns shows for each user namespace ➊ the "most senior" process with its name
and PID ➋, as well as the user ID and user name to which this user namespace
belongs. The "most senior" process is also termed the "ealdorman"; it is the
topmost *and oldest* process in the process tree that is still attached to the
user namespace.

Please note, that depending on settings, lxkns might show additional groups of
processes attached to the user namespace when they hav different cgroup
controllers set.

> [!NOTE] The init process with PID&nbsp;1 is always shown first, while all
> other "tenants" are sorted by their process names. Additionally, PID&nbsp;1 is
> marked with an orange crown icon.

## Initial Namespaces

The namespaces ➌ created at system start and attached to PID&nbsp;1 are called
"initial namespaces". These are visually marked by dashed borders to make them
easily spottable.

## Shared Namespaces

Other tenants ➍ might either use some or all of the existing namespaces ("shared
namespaces") or newly created namespaces instead. In case a namespace is – what
we call – "shared" it is shown washed out. Such reused namespaces can be hidden
in the [settings](#settings).

## Containers in Containers

When a process is the top-level process inside a container, **lxkns**
additionally shows the container name ➎.

![containers in containers](_images/all-namespaces-view-dind.png ':class=framedscreenshot')

In our case of ➎ here, the additional "[elegant_haslett]:" annotation signals
that the container named "pinned-canary" is _inside_ the container ➏
"elegant_haslett". This container ➏ then is used as the Development Container of
a Github Codespace.
