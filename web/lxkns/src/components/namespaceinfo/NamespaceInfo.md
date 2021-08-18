### Hierarchical Namespaces

Without further customization, this component renders user namespace information
consisting of the following pieces of data:

- namespace type and ID,
- number of child namespaces (since this is a hierarchical namespace type),
- leader process information (process name and PID),
- user owning the user namespace (user name and UID).

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceInfo
    namespace={{
      ...initProc.namespaces.user,
      ealdorman: {
        ...initProc,
        cgroup: "/world/domination",
      },
    }}
  />
</ComponentCard>;
```

The following example renders PID namespace information, which is similar to
user namespace information, except that it doesn't show any owner information
(please do not confuse here the owner information with information about the
owning user namespace).

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceInfo namespace={initProc.namespaces.pid} />
</ComponentCard>;
```

### Hidden Intermediate Hierarchical Namespaces

Hierarchical namespaces may be sitting somewhere in the hierarchy where they
don't have any attached processes, but child namespaces. We call them "hidden"
in the sense that they do not appear anywhere in the virtual filesystem and can
only be found via the `NS_GET_PARENT` `ioctl()`.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceInfo
    namespace={{
      ...initProc.namespaces.pid,
      reference: [""],
      ealdorman: null,
      leaders: [],
    }}
  />
</ComponentCard>;
```

Hidden hierarchical namespaces differ from fd-referenced or bind-mounted
namespaces without attached processes; such non-hidden namespaces are rendered
as follows instead, much like any other non-hidden namespace.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceInfo
    namespace={{
      ...initProc.namespaces.pid,
      reference: ["/proc/1/ns/mnt", "/run/mnt/foobar"],
      ealdorman: null,
      leaders: [],
    }}
  />
</ComponentCard>;
```

### Flat Namespaces

Render a flat (network) namespace, which in consequence lacks any details of
children – it has none.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceInfo namespace={initProc.namespaces.net} />
</ComponentCard>;
```
