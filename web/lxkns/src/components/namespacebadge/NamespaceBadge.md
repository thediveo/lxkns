At the moment, the Linux kernel defines the following types of namespaces, which
lxkns renders as following:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { NamespaceType } from "models/lxkns";
import { initProc } from 'models/lxkns/mock';

<>
  {Object.values(NamespaceType).sort().map((nstype, idx) => [
    idx > 0 && <br/>,
    <div>{nstype} namespace badge:</div>,
    <ComponentCard>
      <NamespaceBadge namespace={{
          ...initProc.namespaces.pid,
          type: nstype,
      }} />
    </ComponentCard>
  ])}
</>
```
