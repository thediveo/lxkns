This example renders two `DrawerLinkItem`s, the second being an `avatar` item.

```tsx
import { BrowserRouter as Router } from "react-router-dom";
import { Box } from "@mui/material";
import HomeIcon from "@mui/icons-material/Home";

// Need to use a non-existing item path (route) in order to avoid the items
// showing as being the active route with an annoying background.
<Router>
  <Box width="20em" m={1}>
    <DrawerLinkItem key="home" label="Home" icon={<HomeIcon />} path="/grumpfelwurtzsch" />
  </Box>
  <Box width="20em" m={1}>
    <DrawerLinkItem key="home2" avatar label="Home" icon={<HomeIcon />} path="/grumpfelwurtzsch" />
  </Box>
</Router>;
```
