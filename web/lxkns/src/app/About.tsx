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

import Typography from '@material-ui/core/Typography'
import Icon from '@material-ui/core/Icon'

import extlink from 'components/extlink'

import LxknsIcon from "./lxkns.svg"
import version from '../version'

// Render information about this web application.
export const About = () => (<>
    <Typography variant="h4">
        <Icon><img src={LxknsIcon} alt="" /></Icon>
            Linux Kernel Namespaces Discovery App
        </Typography>

    <Typography variant="body2" paragraph={true}>
        app version {version}
    </Typography>

    <Typography variant="body2" paragraph={true}>
        GitHub: {extlink('https://github.com/thediveo/lxkns', 'thediveo/lxkns', true)}
        <br />
        license: {extlink('https://www.apache.org/licenses/LICENSE-2.0', 'Apache License 2.0', true)}
    </Typography>

    <Typography variant="body1" paragraph={true}>
        This app displays all Linux-kernel
        {extlink('https://man7.org/linux/man-pages/man7/namespaces.7.html', 'namespaces', true)}
        discovered inside a Linux host. Namespaces can be either displayed in
        an all-inclusive view (the "home" view), or for each type of namespace
        individually.
        </Typography>

    <Typography variant="body1" paragraph={true}>
        Namespaces (which, ironically, are <i>unnamed</i>) partition certain kinds of
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
        The "home" display is organized along the hierarchy of user namespaces.
        Namespaces of the other types are then shown beneath the particular user namespace
        which is owning them. "Owning a namespace" here means that a namespace was
        created by a process while the process was attached to that specific user
        namespace.
    </Typography>

    <Typography variant="body1" paragraph={true}>
        The type-specific namespace views are slightly different: first, they
        show only namespaces of a single type. Then, they are mostly flat views,
        except for PID and user namespace views. In case there are multiple
        bunches of processes, each bunch with its own cgroup controller, these
        are then shown below the namespaces they are joined to. In case all
        processes joined to a specific namespace are in the same cgroup, the
        "most senior" process of them will be shown right next to the namespace
        information instead.  
    </Typography>

    <Typography variant="body1" paragraph={true}>
        Please note that not all namespaces need to have processes joined to
        them. Namespaces can also exist because they have been bind-mounted
        to some path in the file system or are still referenced by an open
        file descriptor of a process (despite which isn't joined to the
        namespace). Finally, PID and user namespaces can be "hidden" somewhere
        between other PID or user namespaces in the hierarchy, but without any
        process joined to them – this is especially noted in the display.
    </Typography>
</>)

export default About
