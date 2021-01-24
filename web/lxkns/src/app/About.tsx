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

import Box from '@material-ui/core/Box'
import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableContainer from '@material-ui/core/TableContainer'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'
import Paper from '@material-ui/core/Paper'

import Typography from '@material-ui/core/Typography'
import Icon from '@material-ui/core/Icon'

import { ExtLink } from 'components/extlink'

import LxknsIcon from "./lxkns.svg"
import version from '../version'
import { Namespace } from 'models/lxkns'
import { NamespaceInfo } from 'components/namespaceinfo'

const NamespaceExample = ({ type, initial }: { type: string, initial?: boolean }) =>
    <NamespaceInfo namespace={{
        nsid: 4026531837,
        type: type,
        ealdorman: {},
        initial: initial,
        parent: null,
        children: [],
    } as Namespace} noprocess={true} />

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
        GitHub: <ExtLink href="https://github.com/thediveo/lxkns">thediveo/lxkns</ExtLink>
        <br />
        license: <ExtLink href="https://www.apache.org/licenses/LICENSE-2.0">Apache License 2.0</ExtLink>
    </Typography>

    <Typography variant="body1" paragraph={true}>
        This app displays all Linux-kernel
        <ExtLink href="https://man7.org/linux/man-pages/man7/namespaces.7.html">namespaces</ExtLink>
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

    <Box m={2}>
        <TableContainer component={Paper}>
            <Table size="small">
                <TableHead>
                    <TableRow>
                        <TableCell>Type</TableCell>
                        <TableCell>Partion</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    <TableRow>
                        <TableCell><NamespaceExample type="cgroup" /></TableCell>
                        <TableCell>partitions the root directory of cgroup controllers in the file system.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="ipc" /></TableCell>
                        <TableCell>partitions SYSV inter-process communication and
                            POSIX message queues.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="mnt" /></TableCell>
                        <TableCell>partitions file system mount points.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="net" /></TableCell>
                        <TableCell>partitions network stacks with their interfaces,
                            addresses, ports, et cetera.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="pid" /></TableCell>
                        <TableCell>partitions process identifiers (PIDs); this is a
                        hierarchical namespace type, so processes in a parent PID namespace
                            see all processes in child PID namespaces, but not vice versa.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="user" /></TableCell>
                        <TableCell>partitions user and group identifiers (UIDs, GIDs);
                        this is a nested namespace type, so a particular user namespace
                            is affected by the chain of parent user namespaces.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="uts" /></TableCell>
                        <TableCell>partitions the host name and NIS domain name.</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell><NamespaceExample type="time" /></TableCell>
                        <TableCell>since Linux 5.6, partitions the boot and monotonic
                            clocks.</TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </TableContainer>
    </Box>

    <Typography variant="body1" paragraph={true}>
        The so-called "initial namespaces" are created automatically by the Linux kernel itself when
        it starts. Initial namespaces are represented using a dashed border like this in order to make
        them easily identifiable: <NamespaceExample type="net" initial={true} />.
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
