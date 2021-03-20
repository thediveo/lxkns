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

import React from 'react'
import { mount, mountHook } from '@cypress/react'
import { localStorageAtom } from './persistentsettings'
import { Provider as StateProvider, useAtom, Atom, WritableAtom } from 'jotai'

const AtomChanger = ({ atom }) => {
    const [value, setValue] = useAtom(atom)

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        setValue(event.target.value)
    }

    return (<>
        <div id="atom">{`.${value.toString()}.`}</div>
        <select id="changer" name="value" defaultValue={'foo'} onChange={handleChange}>
            <option value={'foo'}>foo</option>
            <option value={'bar'}>bar</option>
        </select>
    </>)
}

const StorageAtomValue = ({ atom }: { atom: Atom<any> }) => {
    const [value] = useAtom(atom)
    return <div id="atom">{`.${value.toString()}.`}</div>
}

const atomValueEquals = <T,>(atom: Atom<T>, value: T) => { // https://stackoverflow.com/a/32697733
    mount(<StateProvider><StorageAtomValue atom={atom} /></StateProvider>)
    cy.waitForReact()
    cy.get('#atom').contains(`.${value.toString()}.`)
}

describe('persistentsettings', () => {

    beforeEach(() => {
        localStorage.clear()
    })

    it('defaults', () => {
        atomValueEquals(localStorageAtom('boolean-atom', true), true)
        atomValueEquals(localStorageAtom('number-atom', 42), 42)
        atomValueEquals(localStorageAtom('string-atom', 'foobar'), 'foobar')
    })

    it('does not persists defaults', () => {
        atomValueEquals(localStorageAtom('number-atom', 42), 42)
        atomValueEquals(localStorageAtom('number-atom', 666), 666)
    })

    it('picks up persistent values', () => {
        [
            ['false', false, true],
            ['true', true, false],
            ['off', false, true],
            ['on', true, false],
            ['"off"', false, true],
            ['"on"', true, false],
            ['foobar', true, true], // uses default value
            ['"foobar"', true, true] // dito.
        ].forEach((test) => {
            localStorage.setItem('boolean-atom', test[0] as string)
            atomValueEquals(localStorageAtom('boolean-atom', test[2] as boolean), test[1] as boolean)
        })

        localStorage.setItem('number-atom', '666')
        atomValueEquals(localStorageAtom('number-atom', 42), 666)
    })

    it('persists changes', () => {
        const atom = localStorageAtom('some-atom', '')
        mount(<StateProvider><AtomChanger atom={atom} /></StateProvider>)
        cy.waitForReact()
        cy
            .get('#atom').contains('..')
            .get('#changer').select('bar')
            .get('#atom').contains('.bar.')
            .then(() => {
                expect(localStorage.getItem('some-atom')).to.equal('"bar"')
            })
            .get('#changer').select('foo')
            .then(() => {
                expect(localStorage.getItem('some-atom')).to.equal('"foo"')
            })
    })

})
