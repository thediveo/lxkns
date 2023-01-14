import data from './mockdata.json'
import { fromjson } from '../fromjson'

export const discovery = fromjson(data)

export const initProc = discovery.processes['1']

export const fakeBindmountedIpc = {
    nsid: initProc.namespaces.ipc.nsid,
    type: initProc.namespaces.ipc.type,
    owner: initProc.namespaces.ipc.owner,
    reference: ['/proc/1/ns/mnt', '/run/snapd/ns/foobar.ipc'],
}

export const fakeBindmountedIpcElsewhere = {
    nsid: initProc.namespaces.ipc.nsid,
    type: initProc.namespaces.ipc.type,
    owner: initProc.namespaces.ipc.owner,
    reference: ['/proc/1234/ns/mnt', '/run/snapd/ns/foobar.ipc'],
}

export const fakeFdIpc = {
    nsid: initProc.namespaces.ipc.nsid,
    type: initProc.namespaces.ipc.type,
    owner: initProc.namespaces.ipc.owner,
    reference: ['/proc/666/fd/666'],
}

export const fakeHiddenPid = {
    nsid: initProc.namespaces.pid.nsid,
    type: initProc.namespaces.pid.type,
    owner: initProc.namespaces.pid.owner,
    reference: [],
}
