```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { MountpointPath } from "components/mountpointpath";
import { mountpoint } from "components/mountpointinfo/fakedata";
const hiddenmountpoint = {
  ...mountpoint,
  hidden: true,
};

<>
  <p>visible mount point</p>
  <ComponentCard>
    <MountpointPath mountpoint={mountpoint} />
  </ComponentCard>

  <p>visible mount point, drum="always"</p>
  <ComponentCard>
    <MountpointPath drum="always" mountpoint={mountpoint} />
  </ComponentCard>

  <p>hidden mount point</p>
  <ComponentCard>
    <MountpointPath mountpoint={hiddenmountpoint} />
  </ComponentCard>

  <p>hidden mount point, drum="never"</p>
  <ComponentCard>
    <MountpointPath drum="never" mountpoint={hiddenmountpoint} />
  </ComponentCard>
</>;
```
