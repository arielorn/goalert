// ***********************************************************
// This example support/index.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:

import './alert'
import './service'
import './ep'
import './rotation'
import './graphql'
import './login'
import './profile'
import './schedule'
import './select-by-label'
import './menu'
import './navitage-to-and-from'
import './page-search'
import './page-action'
import './page-nav'
import './page-fab'
import './config'

export * from './util'

import './fail-fast'

// Alternatively you can use CommonJS syntax:
// require('./commands')
