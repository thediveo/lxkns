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

import React, { useEffect, useMemo, useState, useRef } from 'react'

import { useAtom } from 'jotai'

import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'

import Typography from '@mui/material/Typography'
import { TreeView, TreeItem } from '@mui/x-tree-view'

import ProcessInfo from 'components/processinfo'
import { NamespaceInfo } from 'components/namespaceinfo'
import { Busybody, compareBusybodies, compareNamespaceById, compareProcessByNameId, Discovery, isProcess, Namespace, NamespaceMap, NamespaceType, Process, ProcessMap } from 'models/lxkns'
import { Action, EXPANDALL, COLLAPSEALL } from 'app/treeaction'
import { expandInitiallyAtom, showSystemProcessesAtom } from 'views/settings'
import { MountpointInfoModalProvider } from 'components/mountpointinfomodal'
import { TaskInfo } from 'components/taskinfo'


/** Internal helper to filter "system(d)" processes. */
const showProcess = (process: Process, showSystemProcs: boolean) =>
    showSystemProcs ||
    (process.pid > 2 &&
        !(process.cpucgroup.startsWith('/system.slice/') &&
            !process.cpucgroup.startsWith('/system.slice/docker-')) &&
        !process.cpucgroup.startsWith('/init.scope/') &&
        process.cpucgroup !== '/user.slice')

/**
 * Searches for sub-processes of a given process which are still in the same
 * PID namespace as the process we started from, but which have different
 * controllers (cgroup paths). Returns a flat list of the next-level sub
 * processes.
 *
 * @param proc process to start the search from.
 */
const findSubProcesses = (proc: Process, nstype: NamespaceType): Process[] => {
    // We'll work only on children which are still in the same namespace, all
    // other children can immediately be filtered out.
    const children = proc.children
        .filter(child => child.namespaces[nstype] === proc.namespaces[nstype])
    // We need to recursively check children which are controlled by the same
    // controller as our process, because a change in the controller might be
    // further down the process tree.
    const subprocs = children
        .filter(child => child.cpucgroup === proc.cpucgroup)
        .map(child => findSubProcesses(child, nstype))
        .flat()
    // Finally return the concatenation of all immediate child processes as
    // well as processes further down the hierarchy with controllers differing
    // to our controller.
    return children
        .filter(child => child.cpucgroup !== proc.cpucgroup)
        .concat(subprocs)
}

/** Returns all leader and sub-leader processes in a namespace. */
const findNamespaceProcesses = (namespace: Namespace) =>
    namespace.leaders.concat(
        namespace.leaders.map(leader => findSubProcesses(leader, namespace.type)).flat())

/**
 * Renders a process and then recursively decends down to find and render
 * deeper processes which still belong to the same type of namespace, yet have
 * a different controller (cgroup path).
 *
 * @param proc process
 * @param nstype type of namespace confining the search for further
 * sub-processes still considered to be confined in the same namespace.
 */
const controlledProcessTreeItem = (proc: Process, nstype: NamespaceType, showSystemProcesses: boolean): JSX.Element | JSX.Element[] => {

    const children = findSubProcesses(proc, nstype)
        .sort(compareProcessByNameId)
        .map(child => controlledProcessTreeItem(child, nstype, showSystemProcesses))
        .flat()

    // Special case: this is the only leader process in the namespace and there
    // are no (further) sub-processes with different controllers.
    const hideMe = proc.namespaces[nstype]?.leaders.length === 1 &&
        proc === proc.namespaces[nstype]?.ealdorman

    return (
        (!hideMe && showProcess(proc, showSystemProcesses) &&
            <TreeItem
                className="controlledprocess"
                key={proc.pid}
                nodeId={proc.pid.toString()}
                label={<ProcessInfo process={proc} />}
            >{children}</TreeItem>
        ) || children || null
    )
}

/**
 * Renders a single namespace node including processes (and loose threads)
 * joined to this namespace, and child namespaces in case of hierarchical
 * namespaces. Instead of just dumping a rather useless plain process tree, this
 * component renders only leaders and then additionally only those sub-processes
 * in different cgroups. Loose threads are simply listed (but not grouped in any
 * way).
 *
 * @param namespace namespace information.
 */
const NamespaceTreeItem = (
    namespace: Namespace,
    processes: ProcessMap,
    showSystemProcesses: boolean,
    DetailsFactory?: NamespaceProcessTreeDetailFactory
) => {

    // For later display we want to know if there is only one loose threads and
    // no leaders so that we don't render the loose thread twice...
    const loosethreads = namespace.ealdorman === null && namespace.loosethreads.length === 1
        ? [] : namespace.loosethreads

    // Get the leader processes and maybe some sub-processes (in different
    // cgroups), all inside this namespace. Please note that if there is only a
    // single leader process, then it won't show up -- it has already been
    // indicated as part of the namespace information and thus
    // controlledProcessTreeItem will drop it.
    //
    // Also get the loose threads, if any.
    const busybodies = (namespace.leaders as Busybody[])
        .concat(loosethreads)
        .sort(compareBusybodies)
        .map(busybody => isProcess(busybody)
            ? controlledProcessTreeItem(busybody, namespace.type, showSystemProcesses)
            : <TreeItem
                className="controlledtask"
                key={busybody.tid}
                nodeId={busybody.tid.toString()}
                label={<TaskInfo task={busybody} />}
            />
        )
        .flat()

    // In case of hierarchical namespaces also render the child namespaces.
    const childnamespaces = namespace.children ?
        namespace.children.map(childns => NamespaceTreeItem(childns, processes, showSystemProcesses, undefined)) : []

    return <TreeItem
        className="namespace"
        key={namespace.nsid}
        nodeId={namespace.nsid.toString()}
        label={<NamespaceInfo namespace={namespace} processes={processes} />}
    >{[
        ...busybodies,
        ...childnamespaces,
        ...(DetailsFactory ? [<DetailsFactory key="42" namespace={namespace} />] : [])
    ]}</TreeItem>
}

/**
 * The properties passed to a component for rendering the details of a
 * namespace.
 */
export interface NamespaceProcessTreeDetailComponentProps {
    /** namespace to render more details of. */
    namespace: Namespace
}

/**
 * Factory for returning components to render the details of a particular
 * namespace.
 */
export type NamespaceProcessTreeDetailFactory = (props: NamespaceProcessTreeDetailComponentProps) => JSX.Element

export interface NamespaceProcessTreeTreeDetails {
    factory: NamespaceProcessTreeDetailFactory
    collapseAll?: (namespaces: NamespaceMap) => string[]
    expandAll?: (namespaces: NamespaceMap) => string[]
}

export interface NamespaceProcessTreeProps {
    type?: string
    action: Action
    discovery: Discovery
    details?: NamespaceProcessTreeTreeDetails
}

/**
 * Component `NamespaceProcessTree` renders a tree of namespaces of a specific
 * type only, with their contained processes. Here, contained processes are not
 * only leader processes in a namespace, but also (grand) child processes within
 * the same namespace, but with different controllers (cgroup paths). In case of
 * non-hierarchical namespace types, the namespace tree is flat.
 *
 * @param type type of namespace.
 */
export const NamespaceProcessTree = ({
    type,
    action,
    discovery,
    details
}: NamespaceProcessTreeProps) => {

    const nstype = type as NamespaceType || NamespaceType.pid

    // System process filter setting.
    const [showSystemProcesses] = useAtom(showSystemProcessesAtom)

    // Expand new nodes?
    const [expandInitially] = useAtom(expandInitiallyAtom)

    // Previous discovery information, if any.
    const previousDiscovery = useRef({ namespaces: {}, processes: {} } as Discovery)

    // Tree node expansion is a component-local state. We need to also use a
    // reference to the really current expansion state as for yet unknown
    // reasons setExpanded() will pass stale state information to its reducer.
    const [expanded, setExpanded] = useState([] as string[])
    const currExpanded = useRef([] as string[])

    // Remember the current tree node expansion state in order to be able to
    // later determine how to correctly deal with discovery updates and node
    // expansion.
    useEffect(() => {
        currExpanded.current = expanded
    }, [expanded])

    // Trigger an action when the action "state" changes; we are looking only at
    // the action itself, ignoring its mutation counter, as we need to add noise
    // to the commands in order to make state changes trigger. Oh, well, bummer.
    useEffect(() => {
        switch (action.action) {
            case EXPANDALL: {
                // expand all namespaces as well as all confined processes.
                const allns = Object.values(discovery.namespaces)
                    .filter(ns => ns.type === nstype)
                const allnsids = allns.map(ns => ns.nsid.toString())
                const allprocids = allns.map(ns => findNamespaceProcesses(ns))
                    .flat()
                    .map(proc => proc.pid.toString())
                setExpanded(allnsids.concat(
                    allprocids,
                    details?.expandAll ? details.expandAll(discovery.namespaces) : []))
                break
            }
            case COLLAPSEALL: {
                // collapse everything except for the root namespaces.
                const allrootnsids = Object.values(discovery.namespaces)
                    .filter(ns => ns.type === nstype && ns.parent == null)
                    .map(ns => ns.nsid.toString())
                setExpanded(allrootnsids.concat(
                    details?.collapseAll ? details.collapseAll(discovery.namespaces) : []))
                break
            }
        }
    }, [action, nstype, discovery, details])

    // Whenever the discovery changes, we want to update the expansion state of
    // the newly arrived namespaces. We default to newly seen "root" namespaces
    // being expanded to show their processes (if there are multiple leader and
    // sub-leader processes).
    useEffect(() => {
        // Unfortunately, the material-ui tree maintains only a single list of
        // expanded node ids. In order to not touch the expanded/collapsed state
        // of existing nodes we thus first need to find out which nodes actually
        // are new (and for hierarchical namespaces we don't care about
        // hierarchy).
        const previousNamespaces = previousDiscovery.current.namespaces as NamespaceMap
        const oldNamespaceIds = Object.values(previousNamespaces)
            .filter(ns => ns.type === nstype)
            .map(ns => ns.nsid)
        // We want to expand only new namespaces and never touch their expansion
        // state lateron. First, we select a suitable filter depending on the
        // type of namespaces to be rendered and the setting for expanding
        // "top-level" namespace nodes.
        const expansionCandidateFilter: (ns: Namespace) => boolean =
            expandInitially
                ? (ns: Namespace) => (ns.type === nstype)
                : ((nstype === 'user' || nstype === 'pid')
                    ? (ns: Namespace) => (ns.type === nstype && (ns.initial || false))
                    : (ns: Namespace) => (ns.type === nstype))
        const expandingNamespaces = Object.values(discovery.namespaces)
            .filter(expansionCandidateFilter)
            .filter(ns => !oldNamespaceIds.includes(ns.nsid))
        // Finally update the expansion state of the tree; this must include the
        // already expanded nodes (state), so that already expanded nodes don't
        // collapse on the next refresh.
        const expandNodeIds = expandingNamespaces.map(ns => ns.nsid.toString())
        setExpanded(currExpanded.current.concat(expandNodeIds))
        previousDiscovery.current = discovery
    }, [nstype, discovery, expandInitially])

    // Whenever the user clicks on the expand/close icon next to a tree item,
    // update the tree's expand state accordingly. This allows us to explicitly
    // "take back control" (ha ... hah ... HAHAHAHA!!!) of the expansion state
    // of the tree.
    const handleToggle = (event: unknown, nodeIds: string[]) => {
        setExpanded(nodeIds)
    }

    // Memorize the tree items, so we don't need to rerender them unless we've
    // got new discovery data or the display filter changes; this avoids
    // rerendering the tree contents when changing the "expanded" tree state.
    const treeItemsMemo = useMemo(() => (
        Object.values(discovery.namespaces)
            .filter(ns => ns.type === nstype && ns.parent == null)
            .sort(compareNamespaceById)
            .map(ns => NamespaceTreeItem(ns, discovery.processes, showSystemProcesses, (details && details.factory) || undefined))
    ), [discovery, showSystemProcesses, nstype, details])

    return (
        (treeItemsMemo.length &&
            <MountpointInfoModalProvider namespaces={discovery.namespaces}>
                <TreeView
                    className="namespacetree"
                    disableSelection={true}
                    onNodeToggle={handleToggle}
                    defaultCollapseIcon={<ExpandMoreIcon />}
                    defaultExpandIcon={<ChevronRightIcon />}
                    expanded={expanded}
                >{treeItemsMemo}</TreeView>
            </MountpointInfoModalProvider>
        ) || (Object.keys(discovery.namespaces).length &&
            <Typography variant="body1" color="textSecondary">
                this Linux system doesn&apos;t have any &ldquo;{nstype}&rdquo; namespaces
            </Typography>
        ) || (
            <Typography variant="body1" color="textSecondary">
                nothing discovered yet, please refresh
            </Typography>
        )
    )
}

export default NamespaceProcessTree
