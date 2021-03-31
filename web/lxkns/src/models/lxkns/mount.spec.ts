// Copyright 2021 Harald Albrecht.
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

import { MountPath, MountPoint, starterDir, insertCommonChildPrefixMountPaths, unescapeMountPath } from './mount'

describe('mount helpers', () => {

    it('unescapes mount paths', () => {
        expect(unescapeMountPath('foo\\040bar\\057')).to.equal('foo bar/')
        expect(unescapeMountPath('foo\\bar')).to.equal('foo\\bar')
    })

    it('gets starter directories', () => {
        expect(starterDir('/abc/def')).to.equal('abc')
        expect(starterDir('abc/def')).to.equal('abc')
        expect(starterDir('/abc')).to.equal('abc')
        expect(starterDir('abc')).to.equal('abc')
        expect(starterDir('/')).to.equal('')
        expect(starterDir('')).to.equal('')
    })

    it('inserts common child prefix mount paths', () => {
        // note: '/12/c' is invalid in this simplified test iff it is
        // accompanied by its child mount paths. This test works on only on
        // exactly one level of mount path parent with its next level children.
        const chmp = ['/12/a', '/12/b', '/12/c/a/b/c', '/12/c/a/b/d', '/12/d']
            .map((p, idx) => {
                return ({
                    path: p,
                    mounts: [{ mountid: idx+1, mountpoint: p } as MountPoint],
                    children: [],
                } as MountPath)
            })
        const mproot = {
            path: '/12',
            mounts: [{mountpoint: '/12'} as MountPoint],
            children: chmp,
        } as MountPath
        insertCommonChildPrefixMountPaths(mproot)

        expect(mproot.children.map(mp => mp.path))
            .to.have.members(['/12/a', '/12/b', '/12/c', '/12/d'])
        const mpc = mproot.children.filter(mp => mp.path === '/12/c')
        expect(mpc).to.have.length(1)
        expect(mpc[0].children.map(mp => mp.path))
            .to.have.members(['/12/c/a']) // exactly one, so we know the recursion works.
        expect(mpc[0].children[0].children.map(mp => mp.path))
            .to.have.members(['/12/c/a/b'])
    })

})
