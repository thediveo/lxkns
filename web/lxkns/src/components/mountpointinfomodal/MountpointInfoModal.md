Click on the button to show a modal information dialog with mount point details.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { Button, Tooltip } from "@mui/material";
import {
  MountpointInfoModalProvider,
  useMountpointInfoModal,
} from "components/mountpointinfomodal";
import { mountpoint } from "components/mountpointinfo/fakedata";

const Component = () => {
  const setMountpoint = useMountpointInfoModal();

  return (
    <Tooltip title="click to open Mountpoint dialog">
      <>🖝<Button
        variant="outlined"
        color="primary"
        onClick={() => setMountpoint(mountpoint)}
      >
        ...
      </Button>🖜</>
    </Tooltip>
  );
};

<MountpointInfoModalProvider>
  <Component />
</MountpointInfoModalProvider>;
```
