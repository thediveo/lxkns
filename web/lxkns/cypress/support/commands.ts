/// <reference types="cypress" />
import React from 'react'
import { mount } from 'cypress/react18'
import 'cypress-react-selector'

Cypress.Commands.add('mount', (component: React.ReactNode, options) => {
  // Wrap any parent components needed
  // ie: return mount(<MyProvider>{component}</MyProvider>, options)
  return mount(component, options)
})