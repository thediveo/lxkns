Welcome to the meta level! `ComponentCard`s are a nifty trick to quickly see
what amount of space components take when rendered. And this without the need to
open the web developer's console and then picking around the page, hoping to
correctly hit the component. So let's render a simple text inside a
`ComponentCard`, and all this inside an outher ~`InceptionCard`~, _erm_,
`ComponentCard`.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";

<ComponentCard>
  <ComponentCard>
    Doh!
  </ComponentCard>
</ComponentCard>;
```
