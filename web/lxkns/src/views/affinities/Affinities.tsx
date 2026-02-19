// Copyright 2026 Harald Albrecht.
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

import { Box } from '@mui/material'
import type { TreeAPI } from 'app/treeapi'
import AffinityTree from 'components/affinitytree'
import type { Discovery } from 'models/lxkns'


export interface AffinitiesProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /* lxkns discovery data */
    discovery: Discovery
}

/**
 * The `Affinities` component renders a tree depicting which tasks and processes
 * are runnable on which logical CPUs. The necessary information is already part
 * of the discovery information.
 * 
 * - the top-most level consists of the logical CPUs (numbers) for which we
 *   found tasks and processes.
 * 
 * - immediately below are PID1 and PID2 nodes, where PID2 contains kernel
 *   threads (which are actually seen as processes) and PID1 contains
 *   "user-space" tasks and processes other than kernel threads.
 * 
 * - tasks and processes on a particular logical CPU are always shown in the
 *   context of their ancestry processes; even if some ancestors aren't on this
 *   particular CPU. Such ancestors are rendered in the "disabled" muted color
 *   as to visually differentiate them from tasks and processes that are
 *   actually affine to the CPU where they are rendered.
 * 
 * - the task group leader represents the process itself: it is thus never
 *   rendered as a task node but always as a process node.
 * 
 * - non-group leader tasks are only rendered if their affinities differ from
 *   the affinities of their task group leader.
 * 
 * - a task or process that is affine to only a single logical CPU shows a pin
 *   as a visual clue.
 */
export const Affinities = ({ apiRef, discovery }: AffinitiesProps) => (
    <Box pl={1}>
        <AffinityTree discovery={discovery} apiRef={apiRef }/>
    </Box>
)
