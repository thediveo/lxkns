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

import { Typography } from "@material-ui/core";
import Box from '@material-ui/core/Box';
import Icon from '@material-ui/core/Icon';

import extlink from 'components/extlink';

import LxknsIcon from "./lxkns.svg";
import version from './version';

// Render information about this web application.
export const About = () => (<Box p={1}>
    <Typography variant="h4">
        <Icon><img src={LxknsIcon} alt="" /></Icon>
            Linux Kernel Namespaces Discovery App
        </Typography>

    <Typography variant="body2" paragraph={true}>
        app version {version} / 
        {extlink('https://www.apache.org/licenses/LICENSE-2.0', 'Apache License 2.0', true)}
    </Typography>

    <Typography variant="body1" paragraph={true}>
        This app displays all Linux-kernel
        {extlink('https://man7.org/linux/man-pages/man7/namespaces.7.html', 'namespaces', true)}
        discovered inside a Linux host. Namespaces partition certain kinds of
        OS resources, so that processes that are members of a namespace see
        their own isolated resources. The following types of namespaces are
        currently defined:
    </Typography>

    <Typography variant="body1" paragraph={true}>
        <ul>
            <li><b>cgroup</b>: partitions the root directory of cgroup controllers
                in the file system.
            </li>
            <li><b>ipc</b>: partitions SYSV inter-process communication and
                POSIX message queues.</li>
            <li><b>mount</b>: partitions mount points.</li>
            <li><b>net</b>: partitions network stacks with their interfaces,
                addresses, ports, et cetera.</li>
            <li><b>pid</b>: partitions process identifiers (PIDs); this is a
                hierarchical namespace type, so processes in a parent PID namespace
                see all processes in child PID namespaces, but not vice versa.</li>
            <li><b>user</b>: partitions user and group identifiers (UIDs, GIDs);
                this is a nested namespace type, so a particular user namespace
                is affected by the chain of parent user namespaces.</li>
            <li><b>uts</b>: partitions the host name and NIS domain name.</li>
            <li><b>time</b>: since Linux 5.6, partitions the boot and monotonic
                clocks.</li>
        </ul>
    </Typography>

    <Typography variant="body1" paragraph={true}>
        The display is organized along the hierarchy of user namespaces.
        Namespaces of other types are shown beneath the particular user namespace
        which is owning them. Owning a namespace here means that a namespace was
        created by a process while the process was attached to that specific user
        namespace.
    </Typography>

    <Typography variant="body1" paragraph={true}>
        Find the
        {extlink('https://github.com/thediveo/lxkns', 'thediveo/lxkns', true)}
        project on Github.
    </Typography>
</Box>);

export default About;
