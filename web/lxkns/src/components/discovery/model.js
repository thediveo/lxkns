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

// postprocessDiscovery takes the JSON data returned by the lxkns service API
// "/api/namespaces" endpoint and post-processes it, resolving the namespace and
// process references into JS object reference. This allows for easy navigation
// on the post-processed data model.
export const postprocessDiscovery = (data) => {
    // Process all namespaces and add direct object references instead of
    // indirect references in form of keys/ids.
    Object.values(data.namespaces).forEach(ns => {
        if (ns.type === 'user') { 
            ns.tenants = [];
            ns.children = [];
        } else if (ns.type === 'pid') {
            ns.children = [];
        }
    });
    Object.values(data.namespaces).forEach(ns => {
        // replace leader PIDs with leader process object references ... if
        // there is a list of leader PIDs; otherwise, set an empty array.
        let leaders = [];
        ns['leaders'] && ns.leaders.forEach(leader => {
            if (leader.toString() in data.processes) {
                leaders.push(data.processes[leader.toString()]);
            }
        });
        ns.leaders = leaders;
        // resolve ealdorman, if present; otherwise null.
        ns.ealdorman = (ns['ealdorman'] && data.processes[ns['ealdorman'].toString()]) || null;
        // resolve namspace hierarchy references, if present.
        if (ns.type === 'user' || ns.type === 'pid') {
            ns.parent = (ns['parent'] && data.namespaces[ns.parent.toString()]) || null;
            ns.parent && ns.parent.children.push(ns);
        }
        // resolve ownership, if applicable.
        if (ns['owner']) {
            ns.owner = data.namespaces[ns.owner.toString()];
            ns.owner.tenants.push(ns);
        }
    });
    // Process all, erm, processes and add object references for the hierarchy,
    // making navigation quick and easy.
    Object.values(data.processes).forEach(proc => {
        proc.parent = null;
        proc.children = [];
    });
    Object.values(data.processes).forEach(proc => {
        // Resolve the parent-child relationships.
        if (proc.ppid.toString() in data.processes) {
            proc.parent = data.processes[proc.ppid.toString()];
            proc.parent.children.push(proc);
        }
        // Resolve the attached namespaces relationships.
        for (const [type, nsref] of Object.entries(proc.namespaces)) {
            proc.namespaces[type] = data.namespaces[nsref.toString()]
        }
    });
    // Phew. Done.
    return data;
}

export const namespaceIdOrder = (nsa, nsb) => 
    (nsa.nsid > nsb.nsid) ? 1 : (nsa.nsid < nsb.nsid) ? -1 : 0;

export const namespaceTypeIdOrder = (nsa, nsb) =>
    (nsa.type > nsb.type) ? 1 : (nsa.type < nsb.type) ? -1 : 
        namespaceIdOrder(nsa, nsb);

export const namespaceNameTypeIdOrder = (nsa, nsb) =>
    (nsa.reference > nsb.reference) ? 1 : (nsa.reference < nsb.reference) ? -1 :
        (nsa.type > nsb.type) ? 1 : (nsa.type < nsb.type) ? -1 : 
            namespaceIdOrder(nsa, nsb);

export const processNameIdOrder = (proc1, proc2) => {
    const c = proc1.name.localeCompare(proc2.name);
    return c !== 0 ? c : proc2.pid - proc1.pid;
};
    