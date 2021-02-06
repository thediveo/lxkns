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

import React from 'react'

import { SvgIconProps } from '@material-ui/core'

import { NamespaceType, } from "models/lxkns"
import { namespaceTypeInfo } from './iconmap'


// We extend Material UI's SVG icon properties with a namespace property, from
// which we later can espy the required type of namespace information. 
export interface NamespaceIconProps extends SvgIconProps {
    /** namespace type. */
    type: NamespaceType
}

/**
 * Renders a namespace icon based on the type of the given namespace. Namespace
 * icons are SVG icons.
 */
export const NamespaceIcon = ({type, ...props}: NamespaceIconProps) =>
    type ?
        React.createElement(namespaceTypeInfo[type].icon, props)
        : null
