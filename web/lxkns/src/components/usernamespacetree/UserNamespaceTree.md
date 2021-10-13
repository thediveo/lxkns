Renders the tree of hierarchical user namespaces, and additionally also render
the attached processes with the non-user namespaces these processes are attached
to (in addition to the user namespaces). Oh, the fun of Linux-kernel namespaces.

```tsx
import { UserNamespaceTree } from "components/usernamespacetree";
import { discovery } from "views/help/fakehelpdata";
import { Provider, useAtom } from "jotai";

<Provider>
  <UserNamespaceTree discovery={discovery} action={{ action: "" }} />
</Provider>;
```
