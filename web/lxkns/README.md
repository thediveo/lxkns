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

- `src`
- `static`
- `build`
