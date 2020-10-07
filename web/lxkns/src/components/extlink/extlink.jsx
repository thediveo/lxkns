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

import React from 'react';

import LaunchIcon from '@material-ui/icons/Launch';

const extlink = (href, linktext, spaced) => (<>
    {spaced && ' '}
    <LaunchIcon fontSize="inherit" className="inlineicon" style={{ verticalAlign: 'middle' }} /><a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
    >{linktext}</a>
    {spaced && ' '}
</>);

export default extlink;
