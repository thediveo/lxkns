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

import { useState } from 'react'
import { mount } from '@cypress/react'
import useId from './id'

const Component = () => {

    const [state, setState] = useState('foo')

    const id1 = useId(`${state}-`)
    const id2 = useId('bar-')

    const handleClick = () => {
        setState(`${state}-oh`)
    }

    return (
        <div className='comp' onClick={handleClick}>
            <span>{id1}</span>
            <span>{id2}</span>
            <span>{state}</span>
        </div>
    )

}

describe('id', () => {

    it('creates unique ids', () => {
        mount(<>
            <Component />
            <Component />
        </>)
        cy
            .get('.comp>span').first().as('id1').then((elem) => {
                expect(elem.text()).to.match(/foo-\d+/)
            })
            .get('.comp>span').eq(1).then((elem) => {
                expect(elem.text()).to.match(/bar-\d+/)
                cy.get('@id1').then((id1) => {
                    expect(id1.text().split('-')[1])
                        .not.to.be.equal(elem.text().split('-')[1])
                })
            })
            .get('.comp>span').eq(2).then((elem2) => {
                cy.get('@id1').then((id1) => {
                    expect(elem2.text()).not.to.be.equal(id1.text())
                })
            })
    })

    it('keeps stable unique ids over the lifetime of a component', () => {
        mount(<Component />)
        cy
            .get('.comp>span').first().as('id1')
            .click()
            .get('.comp>span').first().then((elem) => {
                cy.get('@id1').then((id1) => {
                    expect(elem.text()).to.be.equal(id1.text())
                })
            })
            .get('.comp>span').eq(2).contains('foo-oh')
    })

})
