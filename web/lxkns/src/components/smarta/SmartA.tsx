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
import { Link } from 'react-router-dom'

import { ExtLink } from 'components/extlink'


export interface SmartAProps {
    /** hyper reference */
    href: string
    /** children to render inside the hyperlink. */
    children: React.ReactNode
}

/**
 * Renders a hyperlink either as an external link (using the ExtLink component),
 * or a React (DOM) router "internal" Link component, depending on the given
 * href property value. Using the Link component ensures proper app-internal
 * route handling without having to reload the application and thus destroying
 * the any discovery result.
 */
export const SmartA = ({href, children, ...otherprops}: SmartAProps) => {
    try {
        new URL(href)
        return <ExtLink href={href} {...otherprops}>{children}</ExtLink>
    } catch {
        return <Link to={href} {...otherprops}>{children}</Link>
    }
}

export default SmartA
