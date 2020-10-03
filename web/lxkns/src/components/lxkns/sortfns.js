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
