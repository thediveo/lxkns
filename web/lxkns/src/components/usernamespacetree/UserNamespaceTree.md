```tsx
import { UserNamespaceTree } from "components/usernamespacetree";
import { discovery } from "views/help/fakehelpdata";
import { Provider, useAtom } from "jotai";

<Provider>
  <UserNamespaceTree discovery={discovery} action={{ action: "" }} />
</Provider>;
```
