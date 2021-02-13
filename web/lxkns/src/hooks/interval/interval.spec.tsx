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

import { mount, mountHook } from '@cypress/react'
import { useState } from 'react'
import useInterval from './interval'

// Simple functional component to test the useInterval hook while changing the
// interval.
const Ticker = ({ callback }: { callback: () => void }) => {

    const [interval, setInterval] = useState(null)

    useInterval(callback, interval)

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        // oh dear HTML5, selecting the null value gives us the "off" value
        // instead, so we need to map "off" back to a null value. Ouch.
        setInterval(event.target.value !== 'off' ? event.target.value : null)
    }

    return (<>
        <div id="interval">interval: {interval || 'off'}</div>
        <select id="ticker" name="interval" size={5} defaultValue={null} onChange={handleChange}>
            <option value={null}>off</option>
            <option value={1000}>1s</option>
            <option value={5000}>5s</option>
        </select>
    </>)
}

describe('interval', () => {

    it('calls back at regular interval', () => {
        const cb = cy.stub().as('stub')
        // Ouch; always keep Cypress' "asynchronousness" in mind. Mocking the
        // native timers using cy.clock() is done asynchronously, so we have to
        // wait for it to be done using then(). Only then can we proceed to
        // mount out HUT (=hook-under-test), as otherweise the hook would, well,
        // into the native setInterval instead of the mocked one.
        cy.clock().then(() => {
            mountHook(() => useInterval(cb, 1000))
        })
            .tick(500)
            .get('@stub', { timeout: 0 }).should('not.have.been.called')
            .tick(500)
            .get('@stub', { timeout: 0 }).should('have.been.calledOnce')
            .tick(400)
            .get('@stub', { timeout: 0 }).should('have.been.calledOnce')
            .tick(600)
            .get('@stub', { timeout: 0 }).should('have.been.calledTwice')
    })

    it('does not call back for "null" interval', () => {
        const cb = cy.stub().as('stub')
        cy.clock().then(() => {
            mountHook(() => useInterval(cb, null))
        })
            .tick(100000)
            .get('@stub', { timeout: 0 }).should('not.have.been.called')
    })

    it('changes tick, erm, tack', () => {
        const cb = cy.stub().as('stub')
        cy.clock().then(() => {
            mount(<Ticker callback={cb} />)
        })
        cy
            .tick(10000)
            .get('@stub', { timeout: 0 })
            .should('not.have.been.called')

            .get('#ticker').select('5s')
            .get('#interval').contains('interval: 5000')
            .tick(2000)
            .get('@stub', { timeout: 0 }).should('not.have.been.called')
            .tick(3000)
            .get('@stub', { timeout: 0 }).should('have.been.calledOnce')
            .tick(10000)
            .get('@stub', { timeout: 0 }).should('have.been.calledThrice')

            .get('#ticker').select('off')
            .get('@stub', { timeout: 0 }).should('have.been.calledThrice')
            .get('#interval').contains('interval: off')
            .tick(10000)
            .get('@stub', { timeout: 0 }).should('have.been.calledThrice')

            .get('#ticker').select('1s')
            .get('#interval').contains('interval: 1000')
            .tick(2000)
            .get('@stub', { timeout: 0 }).should('have.been.callCount', 5)
            .tick(500)
            .get('@stub', { timeout: 0 }).should('have.been.callCount', 5)

            .get('#ticker').select('off')
            .get('#interval').contains('interval: off')
            .tick(10000)
            .get('@stub', { timeout: 0 }).should('have.been.callCount', 5)
    })

})
