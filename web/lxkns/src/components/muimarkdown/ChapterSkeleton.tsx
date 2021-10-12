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

import { Skeleton, styled, Theme, Typography } from '@mui/material'


interface chapterSkeletonStyleProps {
    theme?: Theme
    rem: number
}

const ChSkeleton = styled('div')(({ theme, rem }: chapterSkeletonStyleProps) => ({
    width: `${rem}rem`,
    '& > :nth-child(1)': { width: `${rem*0.55}rem` },
    '& > :nth-child(2)': { width: `${rem*0.9}rem` },
    '& > :nth-child(3)': { width: `${rem}rem` },
    '& > :nth-child(4)': { width: `${rem*0.7}rem` },
}))


export interface ChapterSkeletonProps {
    /** width of skeleton in 'rem'. */
    rem?: number
}

/**
 * Renders a simple chapter-like skeleton to be used as a fallback while MDX
 * modules are getting lazily loaded.
 */
export const ChapterSkeleton = ({ rem }: ChapterSkeletonProps) => {
    return <ChSkeleton rem={rem || 15}>
        <Typography variant="h4"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
    </ChSkeleton>
}
