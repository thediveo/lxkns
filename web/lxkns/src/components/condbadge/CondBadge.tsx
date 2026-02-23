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

import { Badge, type BadgeProps } from '@mui/material'

export interface CondBadgeProps extends BadgeProps {
    show: boolean
}

/**
 * The `CondBadge` renders its children, optionally with a Badge when the show
 * property is true.
 */
export const CondBadge = ({show, ...props}: CondBadgeProps) => (
    show ? <Badge {...props}>{props.children}</Badge> : props.children
)

export default CondBadge
