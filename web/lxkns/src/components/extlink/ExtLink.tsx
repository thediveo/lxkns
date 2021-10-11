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

import LaunchIcon from '@mui/icons-material/Launch'
import { styled } from '@mui/styles';


const NixWieWegHier = styled('span')(({ theme }) => ({
    // In order to avoid line wraps immediately after the external link
    // icon, wrap (sic!) into a non-wrapping span...
    whiteSpace: 'nowrap',
    // ...and then allow the link text to wrap again.
    '& a': {
        whiteSpace: 'normal',
    },
    // Resize and reposition the external link icon so it fits into the
    // overall text flow and size.
    '& .MuiSvgIcon-root': {
        fontSize: 'inherit',
        verticalAlign: 'middle',
    },
    '& .MuiSvgIcon-root.before': {
        marginRight: '0.1em',
    },
    '& .MuiSvgIcon-root.after': {
        marginLeft: '0.1em',
    },
}))

export interface ExtLinkProps {
    /** href URL. */
    href: string
    /** external link icon placement. */
    iconposition?: 'before' | 'after'
    /** children to render inside the hyperlink. */
    children: React.ReactNode
}

/**
 * Renders an external link together with an "external link" icon before the
 * link text. The external link opens in a new blank tab. Additionally, the link
 * gets set to
 * "[noopener](https://developer.mozilla.org/en-US/docs/Web/HTML/Link_types/noopener)"
 * and
 * "[noreferrer](https://developer.mozilla.org/en-US/docs/Web/HTML/Link_types/noreferrer)"
 * in order to avoid granting the new browsing context access to your single
 * page app and leaking referrer information.
 */
export const ExtLink = ({ href, iconposition, children }: ExtLinkProps) => (
    <NixWieWegHier>
        {iconposition !== 'after' && <LaunchIcon className="before" />}<a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
        >{children}</a>{iconposition === 'after' && <LaunchIcon className="after" />}
    </NixWieWegHier>
)

export default ExtLink
