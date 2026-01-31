// Copyright 2025 Harald Albrecht.
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

import { localStorageAtom } from "utils/persistentsettings"

const themeKey = 'lxkns.theme'
const showSystemProcessesKey = 'lxkns.showsystemprocesses'
const showSharedNamespacesKey = 'lxkns.showsharedns'
const expandInitiallyKey = 'lxkns.expandinitially'
const expandWorkloadInitiallyKey = 'lxkns.expandwlinitially'

export const THEME_USERPREF = 0
export const THEME_LIGHT = 1
export const THEME_DARK = -1
export const themeAtom = localStorageAtom(themeKey, THEME_USERPREF)

export const showSystemProcessesAtom = localStorageAtom(showSystemProcessesKey, false)
export const showSharedNamespacesAtom = localStorageAtom(showSharedNamespacesKey, true)
export const expandInitiallyAtom = localStorageAtom(expandInitiallyKey, false)
export const expandWorkloadInitiallyAtom = localStorageAtom(expandWorkloadInitiallyKey, false)
