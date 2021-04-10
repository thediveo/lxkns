Click on the button to show a modal information dialog with mount point details.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { Button } from "@material-ui/core";
import {
  MountpointInfoModalProvider,
  useMountpointInfoModal,
} from "components/mountpointinfomodal";
import { mountpoint } from "components/mountpointinfo/fakedata";

const Component = () => {
  const setMountpoint = useMountpointInfoModal();

  return (
    <Button variant="outlined" color="primary" onClick={() => setMountpoint(mountpoint)}>
      ...
    </Button>
  );
};

<MountpointInfoModalProvider>
  <Component />
</MountpointInfoModalProvider>;
```
