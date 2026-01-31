# Linux Kernel Namespaces Web App

This part of the repository is home to the `lxkns` web app: a web-browser
application showing the hierarchy of the discovered user namespaces and their
owned non-user namespaces. When served from the `lxkns` discovery server and
loaded into a moderately recent web browser, it will fetch discovery information
from the discovery service and render it in the web browser.

The app itself uses [react](https://reactjs.org) for building its user
interface.

## Deploy and Develop

Whatever you plan to do, open this project in a devcontainer first. This ensures
you have a predefined working environment automatically set up for you.

- to **deploy** as an integral part of the `lxkns` discovery service: in the
  repository's top-level directory do `make deploy` and then navigate to
  `http://localhost:5010`.

- to **develop**:
  - `cd web/lxkns`
  - install all dependencies for this web app: `yarn`.
  - start the wep app development server: `yarn run dev`.
  - switch to storybook: `yarn run storybook`.
  - if you change or add icons (in `icons`, never in `src/icons`): `yarn icons`
