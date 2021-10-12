> **IMPORTANT:** `AppBarDrawer` must be (directly or indirectly) enclosed
> inside a [`(Browser)Router`](https://reactrouter.com/web/api/BrowserRouter)
> component.

```tsx
import { BrowserRouter as Router } from "react-router-dom";
import { Badge, Box, IconButton, List, Typography } from "@mui/material";
import { DrawerLinkItem } from "components/appbardrawer";
import HomeIcon from "@mui/icons-material/Home";
import AnnouncementIcon from "@mui/icons-material/Announcement";
import CachedIcon from "@mui/icons-material/Cached";

<Box
  id="appbardrawerexampleroot"
  height="20ex"
  display="flex"
  flexDirection="column"
>
  <Router>
    <AppBarDrawer
      title={
        <Badge badgeContent={0} color="secondary">
          AwfullApp
        </Badge>
      }
      tools={
        <>
          <IconButton color="inherit">
            <CachedIcon />
          </IconButton>
        </>
      }
      drawerwidth={360}
      drawertitle={
        <>
          <Typography variant="h6" color="textSecondary" component="span">
            AwfullApp
          </Typography>
          <Typography variant="body2" color="textSecondary" component="span">
            &nbsp;0.0.0
          </Typography>
        </>
      }
      drawer={(closeDrawer) => (
        <List onClick={closeDrawer}>
          <DrawerLinkItem
            key="home"
            label="Home"
            icon={<HomeIcon />}
            path="/"
          />
          <DrawerLinkItem
            key="home"
            label="About"
            icon={<AnnouncementIcon />}
            path="/about"
          />
        </List>
      )}
    />
  </Router>
</Box>;
```
