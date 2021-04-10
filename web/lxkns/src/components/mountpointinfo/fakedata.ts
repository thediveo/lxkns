import { MountPoint } from 'models/lxkns/mount'

export const mountpoint: MountPoint = {
    fstype: 'ext2000',
    hidden: false,
    major: 42,
    minor: 123,
    mountid: 2,
    mountoptions: ['foo', 'bar', 'gnampf=uhoh'],
    mountpoint: '/media/fake',
    parentid: 0,
    root: '/',
    source: 'none',
    superoptions: 'super,options',
    tags: {'shared': '123', 'master': '42'},
    children: [],
    parent: null,
    mountnamespace: null,
}
