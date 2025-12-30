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

import React from 'react'

import type { MountPoint } from 'models/lxkns/mount'
import { useContext } from 'react'

const MountpointInfoModalContext = React.createContext<
    undefined | React.Dispatch<React.SetStateAction<MountPoint|undefined>>>(undefined)

export default MountpointInfoModalContext

/**
 * Returns a setter to specify the MountPoint to show information about in a
 * modal dialog.
 */
export const useMountpointInfoModal = () => {
    return useContext(MountpointInfoModalContext)
}
