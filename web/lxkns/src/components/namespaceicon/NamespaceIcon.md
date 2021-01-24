At the moment, the Linux kernel defines the following types of namespaces...

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { NamespaceType } from "models/lxkns";

<>
  {Object.values(NamespaceType).sort().map((nstype, idx) => [
    idx > 0 && <br/>,
    <div>{nstype}:</div>,
    <ComponentCard><NamespaceIcon type={nstype} /></ComponentCard>
  ])}
</>;
```
