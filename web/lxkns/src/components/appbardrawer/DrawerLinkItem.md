This example renders two `DrawerLinkItem`s, the second being an `avatar` item.

```tsx
import { BrowserRouter as Router } from "react-router-dom";
import { Box } from "@material-ui/core";
import HomeIcon from "@material-ui/icons/Home";

<Router>
  <Box width="20em" m={1}>
    <DrawerLinkItem key="home" label="Home" icon={<HomeIcon />} path="/" />
  </Box>
  <Box width="20em" m={1}>
    <DrawerLinkItem key="home2" avatar label="Home" icon={<HomeIcon />} path="/" />
  </Box>
</Router>;
```
