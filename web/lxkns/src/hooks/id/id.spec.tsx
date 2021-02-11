import { mount } from '@cypress/react'
import useId from './id'

const Component = ({ foo }: { foo?: boolean }) => {

    const id1 = useId('foo-')
    const id2 = useId('bar-')

    return (
        <div className='comp'>
            <span>{id1}</span>
            <span>{id2}</span>
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

})
