// Replaces the "npx styleguidist server" command in order to allow for
// asynchronous evaluation of the styleguidist configuration.

import styleguidist from 'react-styleguidist'
import cfgfn from './styleguide.config.js'

//const styleguidist = require('react-styleguidist')
//const cfgfn = require('./styleguide.config.cjs');

(async () => {
    const cfg = await cfgfn()
    styleguidist(cfg).server((err, config) => {
        if (err) {
            console.log(err)
        } else {
            const url = `http://${config.serverHost}:${config.serverPort}`
            console.log(`Listening at ${url}`)
        }
    })
})()
