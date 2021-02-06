// ts-node -O '{"module": "commonjs"}' getmockdata.ts

import http from 'http'
import fs from 'fs'
import md5 from 'md5'

// see: https://www.tomas-dvorak.cz/posts/nodejs-request-without-dependencies/,
// even more simplified.
const httpget = (url: string) => {
    return new Promise((resolve, reject) => {
        const request = http.get(url, (response) => {
            // handle http errors
            if (response.statusCode < 200 || response.statusCode > 299) {
                reject(new Error('status code: ' + response.statusCode));
            }
            const body = [];
            response.on('data', (chunk) => body.push(chunk));
            response.on('end', () => resolve(body.join('')));
        });
        request.on('error', (err) => reject(err))
    })
}

;console.log('fetching discovery data from local lxkns service at :5010...')

const cmd = (c) => {
    return c.split(' ').map((el, idx) => 
        idx < 1 ? el : Math.random().toString(36).slice(2))
        .join(' ')
}

;(async () => {
    let data
    await httpget('http://localhost:5010/api/namespaces')
        .then(result => {data = JSON.parse(result as string)})
        .catch(fail => {
            console.log('error calling lxkns discovery REST API:', fail)
            process.exit(1)
        })
    // render non-root user names anonymous
    const anonymous = Math.random().toString(36).slice(2)
    Object.values(data['namespaces']).forEach(netns => {
        if (netns['user-name'] !== 'root') {
            netns['user-name'] = 'user' + md5(anonymous + netns['user-name'])
        }
    })

    Object.values(data['processes']).forEach(proc => {
        proc['cmdline'] = proc['cmdline'].map((el, idx) => 
            idx < 1 ? cmd(el) : Math.random().toString(36).slice(2))
    })

    const outname = process.argv[2] || 'mockdata.json'
    console.log(`writing ${outname}...`)
    fs.writeFileSync(outname, JSON.stringify(data, null, 4))
})();

console.log('Done.')
