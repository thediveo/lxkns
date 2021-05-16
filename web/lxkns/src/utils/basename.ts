// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

/**
 * Determines this application's basename (path!) from its <base href="..."> DOM
 * element. The basename is "" if unset or "/", otherwise it's the specified
 * basename, but always without any trailing slash. This basename can thus be
 * directly fed into the basename property of a React DOM router component. For
 * this reason, the basename is stripped off of any scheme, host, port, hash,
 * and query elements.
 */
export const basename =
    new URL(
        // get the href attribute of the first base DOM element, falling back to "/"
        // if there isn't one.
        ((document.querySelector('base') || {}).href || '/')
    ).pathname
        // ensure that there is never a trailing slash, and this includes the root
        // itself.
        .replace(/\/$/, '')
