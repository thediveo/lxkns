# Linux Kernel Namespaces Web App

This part of the repository is home to the `lxkns` web app: a web-browser
application showing the hierarchy of the discovered user namespaces and their
owned non-user namespaces. When served from the `lxkns` discovery server and
loaded into a moderately recent web browser, it will fetch discovery information
from the discovery service and render it in the web browser.

The app itself uses [react](https://reactjs.org) for building its user
interface.

## Deploy and Develop

- to **deploy** as an integral part of the `lxkns` discovery service: `make
  deploy` and then navigate to `http://localhost:5010`.

- to **develop**:
  - make sure you have [nodejs](https://nodejs.org) installed.
  - make sure to have yarn [yarn](https://yarnpkg.com/) installed; with `npm`
    installed alongside nodejs: `npm -g install yarn`.
  - install all dependencies for this web app: `yarn install`.
  - start the lxkns service: `run deploy`.
  - start the wep app development server: `yarn start`. 

## Overview

- `src/` – the application source code, organized into a few basic categories.
  Reusable elements are factored out into the `components/` and `hooks/`
  sub-directories. In the following, we won't cover everthings down to the last
  details, but hopefully give you a good view on the **lxkns** react
  application.

  - `app/` – the main application itself, including the `App` component.
    - `App.tsx` – the application's `App` functional component, from which all
      chaos descends.
    - `appstyles.jsx` – defines global styling, supporting both light and dark
      themes. The JSX is used in place of a "traditional" global `.css` file.
    - `treeaction.ts` – the "action" interface between app toolbar buttons and
      the tree view(s).

  - `components/` – potentially reusable components, some more reusable, others
    less.

    - `appbardrawer` – the `AppBarDrawer` component provides apps with the
      usual task bar, as well as a swipeable drawer. This component takes on
      the daunting task of wiring up these things and setting up the standard
      elements, such as the drawer hamburger icon and the app bar title.

    - `discovery` – queries the lxkns discovery API `/api/namespaces` and then
      provides the results via context. Also does some result pre-processing in
      order to allow apps easy and quick navigation on the information model
      using object references, instead of having to look-up IDs all the time.

    - (`elevationscroll` – implements an elevated task bar when the user scrolls
      down. Please note that this component isn't used anymore in this app but
      is kept here for potential reuse in other projects.)

    - `extlink` – the `ExtLink` component suitable for rendering an external
      hyperlink to be opened in a separate tab/window and ensuring a new
      browsing context in order to avoid leaking potential sensitive data and
      referrer data. Also renders a nice external link adornment.

    - `helpviewer` – a multi-chapter help viewer, including chapter navigation.
      Built upon the `muimarkdown` component for rendering MDX content.

    - `muimarkdown` – renders [MDX](https://mdxjs.com/) content following
      Daterial Design typography. MDX is an authorable format that that supports
      JSX inside Markdown documents.

    - `namespacebadge` – renders a namespace "badge" consisting of the type and
      ID of a specific namespace.

    - `namespaceicon` – renders the corresponding type icon for a given
      namespace.

    - `namespaceinfo` – renders a single namespace with additional information,
      such as the ealdorman process and control group information.

    - `namespaceprocesstree` – implements a view into a specific type of
      namespace based on the namespace discovery data.

    - `processinfo` – renders process details, such as process name and PID,
      cgroup path, et cetera. Renders owner information (owner name and ID) for
      user namespaces.

    - `refresher` – provides a one-shot refresh button as well as a refresh
      interval pop-down menu. Automatically interacts with a discovery context.

    - `smarta` – renders an internal (relative) or external (absolute)
      hyperlink, where external hyperlinks get rendered using the `ExtLink`
      component.

    - `usernamespacetree` – implements a view into user namespaces with their
      tenant namespaces.

  - `hooks/` – reusable hooks for functional components.
    - `id` – returns a unique and stable identifier to be used with (rendered)
      HTML elements.
    - `interval` – an interval timer for functional components.

  - `models/lxkns/` – provides the basic lxkns discovery data types for
      `Namespace`, `Process`, et cetera.

    - `mock/` provides mock discovery data.

  - `views/` – contains the main views of the **lxkns** web UI.

    - `about/` – implements the About view, showing copyright and backend service
      version.

    - `help/` – implements the multi-chapter help about **lxkns**.

    - `settings/` – implements the settings views.

  - `utils/` contains a few useful things used across different components, such
    as parsing RGB values in different formats with or without alpha
    transparency.

- `public/` – assets used when serving the application, such as `index.html`
  template, manifest, and app icon(s).

- `build/` – after running `yarn build` in the `web/lxkns` directory, this
  `build/` subdirectory will contain the optimized web app.

- `styleguidist/` – support for the lxkns style guide using [React
  Styleguidist](https://react-styleguidist.js.org/).
