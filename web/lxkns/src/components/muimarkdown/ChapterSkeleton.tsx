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

import { Typography } from '@material-ui/core'
import { Skeleton } from '@material-ui/lab'


export interface ChapterSkeletonProps {
    /** width of skeleton in 'rem'. */
    rem?: number
}

/**
 * Renders a simple chapter-like skeleton to be used as a fallback while MDX
 * modules are getting lazily loaded.
 */
export const ChapterSkeleton = ({ rem }: ChapterSkeletonProps) => {

    rem = rem || 15

    return (<>
        <Typography variant="h4" style={{ width: `${rem / 1.75}rem` }}><Skeleton animation="wave" /></Typography>
        <Typography variant="body1" style={{ width: `${rem}rem` }}><Skeleton animation="wave" /></Typography>
        <Typography variant="body1" style={{ width: `${rem}rem` }}><Skeleton animation="wave" /></Typography>
        <Typography variant="body1" style={{ width: `${rem / 1.2}rem` }}><Skeleton animation="wave" /></Typography>
    </>)
}
