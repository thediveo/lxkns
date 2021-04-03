At the moment, the Linux kernel defines the following types of namespaces, which
lxkns renders as following (in this case also marked as initial namespaces):

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { NamespaceType } from "models/lxkns";
import { initProc } from "models/lxkns/mock";

<>
  {Object.values(NamespaceType)
    .sort()
    .map((nstype, idx) => [
      idx > 0 && <br />,
      <div>{nstype} namespace badge:</div>,
      <ComponentCard>
        <NamespaceBadge
          namespace={{
            ...initProc.namespaces.pid,
            type: nstype,
          }}
        />
      </ComponentCard>,
    ])}
</>;
```

The rendering slightly changes for namespaces which are seen as "shared" between
multiple leader processes: the badges then get washed out and the text fades to
gray 𝅘𝅥🎜. Additionally, the font weight gets changed to a light (300) font
weight.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initProc } from "models/lxkns/mock";

<ComponentCard>
  <NamespaceBadge namespace={initProc.namespaces.cgroup} shared={true} />
</ComponentCard>;
```
