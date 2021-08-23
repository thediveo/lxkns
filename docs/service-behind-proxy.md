# Behind Path-Rewriting Reverse Proxies

The web user interface of the lxkns service is a
[React-based](https://reactjs.org/) so-called [single-page
application](https://en.wikipedia.org/wiki/Single-page_application) (SPA). Now,
serving single-page applications using client-side HTML5 DOM routing behind
path-rewriting reverse proxies always is a challenge.

Of course, as long as you know the exact details where your containerized server
is going to be deployed you might statically set the final base path and build
your React SPA to exactly this configuration. However, if you want to make your
server image more versatile and know the reverse proxy (proxies) in front of
your server will cooperate by telling you the original URL as used by the
client, then things get more flexible.

For lxkns, we use this method: if there is a (rewriting) reverse proxy in front
of our service, it must pass a `X-Forwarded-Uri` HTTP request header with either
the full URL (URI) or at least the absolute path of the resource as originally
requested by a client. This allows our service to determine the "base" path by
comparing the path seen by our service versus the path seen by the first proxy.
This information is then used to dynamically rewrite the `<base href=""/>` from
`index.html` as needed.

In its `public/index.html`, lxkns sets `<base href="%PUBLIC_URL%/"/>` – **please
note the trailing slash!** This will work correctly for development as usual,
where the development server serves from the root.

For the production version we build with `PUBLIC_URL` set to "." (sic!) instead
to "/". This is not a mistake but ensures that all webpack-generated resources
are properly referenced **relative to the (dynamic) base URL**.

Of course, all other web app resources must be referenced using only relative
paths too, including the shortcut/favorite icon, et cetera. There must be no
`%PUBLIC_URL%/` anywhere, except for the `<base />` element.

And the lxkns (REST) API calls must also be relative, too.

In order to make HTML5 DOM routing properly work behind a path-rewriting reverse
proxy the lxkns SPA at runtime picks up its own `<base />` element path and then
passes that on to its DOM router; see `web/lxkns/src/utils/basename.ts` and
`web/lxkns/src/app/App.tsx` for how this is done in lxkns.
