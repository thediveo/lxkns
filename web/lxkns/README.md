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

- `src/` – contains the application top-level elements. Reusable elements are
  factored out into the `components/` and `hooks/` sub-directories.
  - `App.js` – the application's `App` functional component, from which all
    chaos descends. 
  - `components/` – potentially reusable components, some more reusable, others
    less.
    - `appbardrawer` – provides apps with the usual task bar, as well as a
      swipeable drawer. This component takes on the daunting task of wiring up
      these things and setting up the standard elements, such as the drawer
      hamburger icon and the app bar title.
    - `discovery` – queries the lxkns discovery API `/api/namespaces` and then
      provides the results via context. Also does some result pre-processing in
      order to allow apps easy and quick navigation on the information model
      using object references, instead of having to look-up IDs all the time.
    - `elevationscroll` – implements an elevated task bar when the user scrolls
      down.
    - `extlink`
    - `namespace` – renders a single namespace with additional information, such
      as the ealdorman process and control group information.
    - `refresher` – provides a one-shot refresh button as well as a refresh
      interval pop-down menu. Automatically interacts with a discovery context.
    - `usernamespacetree` – implements a view into user namespaces with their
      tenant namespaces.
  - `hooks/` – reusable hooks for functional components.
    - `id` – returns a unique and stable identifier to be used with (rendered)
      HTML elements.
    - `interval` – an interval timer for functional components.
- `static/` –
- `public/` –
- `build/` – after running `yarn build` in the `web/lxkns` directory, this
  `build/` subdirectory will contain the optimized web app.
