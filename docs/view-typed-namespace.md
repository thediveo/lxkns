# Namespace Type Views

The type-specific namespace views show only namespaces for a single specific
type. Most of these views are flat views, except for PID and user namespace
views. The type of namespaces shown is also indicated by the title in the
application bar. The number badge shows the number of namespaces found of this
specific type.

In the screenshot below, the devcontainer "elegant_haslett" (for the "lxkns"
Github codespace) is using the host's PID namespace – therefore, it is listed as
part of the initial PID namespace. In contrast, the "pinned-canary" has been
deployed into its own child PID namespace.

![PID view](_images/pid-namespaces.png ':class=framedscreenshot')

Most types of namespaces are flat without any hierarchy, such as network
namespaces. That is, there aren't network namespaces inside other network
namespaces.

![network view](_images/net-namespaces-compact.png ':class=framedscreenshot')

> [!TIP] If you want to see details inside network namespaces such as network
> interfaces, IP addresses, routes, and more, then we recommend
> [Edgeshark](https://github.com/siemens/edgeshark) – it is built with
> **lxkns**.
