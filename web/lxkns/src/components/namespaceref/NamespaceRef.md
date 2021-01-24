A "hidden" namespace without any reference; as there is no filesystem path
reference, we simply show a ghost:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeHiddenPid } from 'models/lxkns/mock';

<ComponentCard>
    <NamespaceRef namespace={fakeHiddenPid} />
</ComponentCard>
```

A namespace with (only) a file descriptor reference:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeFdIpc } from 'models/lxkns/mock';

<ComponentCard>
    <NamespaceRef namespace={fakeFdIpc} />
</ComponentCard>
```

A namespace with (only) a bind-mount reference:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeBindmountedIpc } from 'models/lxkns/mock';

<ComponentCard>
    <NamespaceRef namespace={fakeBindmountedIpc} />
</ComponentCard>
```
