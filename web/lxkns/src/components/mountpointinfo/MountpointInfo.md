```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { MountpointInfo } from "components/mountpointinfo";
import { mountpoint } from "components/mountpointinfo/fakedata";
const hiddenmountpoint = {
  ...mountpoint,
  hidden: true,
};

<>
  <ComponentCard>
    <MountpointInfo mountpoint={mountpoint} />
  </ComponentCard>
  <ComponentCard>
    <MountpointInfo mountpoint={hiddenmountpoint} />
  </ComponentCard>
</>;
```
