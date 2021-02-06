```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { Provider, useAtom } from "jotai";
import { discoveryRefreshIntervalAtom } from "components/discovery";

const intervals = [
  { interval: null },
  { interval: 666, label: "bad timing" },
  { interval: 1000 },
];

const Interval = () => {
  const [refreshInterval] = useAtom(discoveryRefreshIntervalAtom);
  return <p>INTERVAL: {refreshInterval} (ms)</p>;
};

<Provider>
  <Interval />
  <ComponentCard>
    <Refresher intervals={intervals} />
  </ComponentCard>
</Provider>;
```
