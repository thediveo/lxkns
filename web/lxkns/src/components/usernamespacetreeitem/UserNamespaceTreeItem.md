Renders an user namespace item as part of a user namespace tree. Beside the user
namespace ID this includes further tidbits of information, such as:

- number of child and grandchild user namespaces,
- ealdorman process name and PID, or alternatively container name (and maybe
  project),
- owning (=creating) user ID and user name.

```tsx
import { UserNamespaceTreeItem } from "components/usernamespacetreeitem";
import { discovery } from "views/help/fakehelpdata";
import { Provider, useAtom } from "jotai";

const userns = Object.values(discovery.namespaces)
    .find(ns => ns.type == "user" && ns.initial);

<Provider>
  <UserNamespaceTreeItem namespace={userns} />
</Provider>;
```
