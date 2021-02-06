> **IMPORTANT:** `AppBarDrawer` must be (directly or indirectly) enclosed
> inside a [`(Browser)Router`](https://reactrouter.com/web/api/BrowserRouter)
> component.

```tsx
import { BrowserRouter as Router } from "react-router-dom";
import { Badge, Box, IconButton, List, Typography } from "@material-ui/core";
import { DrawerLinkItem } from "components/appbardrawer";
import HomeIcon from "@material-ui/icons/Home";
import AnnouncementIcon from "@material-ui/icons/Announcement";
import CachedIcon from "@material-ui/icons/Cached";

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
