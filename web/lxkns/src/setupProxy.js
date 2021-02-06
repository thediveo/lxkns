// Ensure that clients requesting the /api REST API path, including subpaths,
// always will get to the real lxkns backend service even if they signal
// accepting "text/html". Otherwise, the simple "proxy": "..." setting in
// package.json would cause any browser address line GET to end up in the proxy
// instead of being handled in the backend service.

const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
    app.use(
        '/api',
        createProxyMiddleware({
            target: 'http://localhost:5010',
            changeOrigin: false,
        })
    );
};
