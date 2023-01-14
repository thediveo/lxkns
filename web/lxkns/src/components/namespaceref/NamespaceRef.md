A "hidden" namespace without any reference; as there is no filesystem path
reference, we simply show a ghost. And finally with an easter egg, just hover
your mouse over the ghost...

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeHiddenPid } from 'models/lxkns/mock';

<ComponentCard>
    <NamespaceRef namespace={fakeHiddenPid} />
</ComponentCard>
```

A namespace with (only) a file descriptor reference, additionally showing process information:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeFdIpc } from 'models/lxkns/mock';

const processes = {
    666: {
        name: "farisee",
    }
};

<ComponentCard>
    <NamespaceRef namespace={fakeFdIpc} processes={processes} />
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

Of course, bind-mounted references can come from other places than the initial mount namespace:

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { fakeBindmountedIpcElsewhere } from 'models/lxkns/mock';

<ComponentCard>
    <NamespaceRef namespace={fakeBindmountedIpcElsewhere} />
</ComponentCard>
```
