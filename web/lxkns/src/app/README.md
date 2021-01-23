# The `app/` Folder

This `app/` folder finally defines the Lxkns App (in form of the `App`
component), by stitching together the main elements of the application. Namely,
the ...

- application bar with app title, drawer menu button, and a few action buttons,
- swipeable drawer for navigation,
- routing for selecting the correct view components based on the current route.
  For details on the views, please refer to the `views/` folder.

Additionally, the App also provides the infrastructure in form of...

- snack bar for showing (error) toasts, such as when the discovery service
  fails.
- `Discovery` component for fetching and providing the latest and greatest lxkns
  discovery results.
- [jotai](https://github.com/pmndrs/jotai) state provider for distributing
  application state to all the different places in the application.

That's it.
