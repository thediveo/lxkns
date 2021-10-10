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

import { Theme, Typography } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { Skeleton } from '@mui/material';


interface chapterSkeletonStyleProps {
    rem: number
}

const useStyles = makeStyles<Theme, chapterSkeletonStyleProps>({
    skeleton: {
        width: props => `${props.rem}rem`,
        '& > :nth-child(1)': { width: props => `${props.rem*0.55}rem` },
        '& > :nth-child(2)': { width: props => `${props.rem*0.9}rem` },
        '& > :nth-child(3)': { width: props => `${props.rem}rem` },
        '& > :nth-child(4)': { width: props => `${props.rem*0.7}rem` },
    }
})

export interface ChapterSkeletonProps {
    /** width of skeleton in 'rem'. */
    rem?: number
}

/**
 * Renders a simple chapter-like skeleton to be used as a fallback while MDX
 * modules are getting lazily loaded.
 */
export const ChapterSkeleton = ({ rem }: ChapterSkeletonProps) => {

    const classes = useStyles({rem: rem || 15})

    return (<div className={classes.skeleton}>
        <Typography variant="h4"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
    </div>)
}
