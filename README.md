# VENI

VENI is two small, independent pieces:

1. `veni.js` is a browser helper for discovering and registering custom-element
   tags already present in a page.
2. `main.go` is an optional visual web-crawler demo. It is useful for trying
   the browser UI locally; it is not the component library and it is not a
   production web-crawler service.

The Go demo and the browser helper do not share an API or a network protocol.
`legacy/` contains parked historical material and is not part of the supported
runtime.

## Browser helper (`veni.js`)

### Requirements and loading

The helper uses browser platform APIs: `Custom Elements`, `HTMLElement`,
`Shadow DOM`, `DOMContentLoaded`, and `Set`. Load it with a normal script:

```html
<script src="/path/to/veni.js"></script>
```

When loaded in a browser it exposes:

- `window.Veni` — the `Veni` class.
- `window.veni` — an automatically initialized singleton.

It has no npm build step and makes no network requests. The script is plain
JavaScript, not a module.

### Minimal example

```html
<template id="t-x-card" data-component="x-card">
  <x-card></x-card>
</template>
<script src="veni.js"></script>
<script>
  class XCard extends HTMLElement {
    connectedCallback() {
      this.textContent = this.textContent || "A card";
    }
  }

  const registry = new Veni({ registry: { "x-card": XCard } });
  registry.init(document);
</script>
```

`init()` scans the current document and the contents of `<template>`
elements. A discovered tag is registered only when its constructor is already
in the instance's `registry`; unknown tags are recorded as pending. Existing
custom elements are not redefined.

The automatically created `window.veni` instance has an empty registry. Code
that loads constructors later can use the same singleton:

```js
veni.define("x-counter", class Counter extends HTMLElement {
  connectedCallback() {
    this.textContent = this.textContent || "0";
  }
});
```

### Actual API

| API | Behavior |
| --- | --- |
| `new Veni({ registry, handlers })` | Creates an independent helper instance. `registry` maps tag names to constructors. |
| `instance.init()` | Scans `document` and template contents; registers matching registry entries. Returns the instance. |
| `instance.pending()` | Returns sorted names discovered but not defined. |
| `instance.autodefine()` | Defines pending names with a small shadow-DOM slot stub. Returns the names defined. |
| `instance.define(name, ctor)` | Defines one custom element. Returns the normalized name or `null` on failure. |
| `instance.register(name, ctor)` | Alias for `define`. |
| `instance.defineAll(map)` | Defines a map of constructors and returns the successful names. |
| `instance.fromTemplate(template, ctor?)` | Derives a valid custom-element name from `data-component`, `data-name`, or an `id`, then defines it. |
| `instance.list()` | Returns the names tracked as defined by this instance. |
| `instance.discover(root)` | Returns currently undefined custom-element tags below `root`. |
| `instance.isDefined(name)` | Checks the browser's global `customElements` registry. |
| `instance.on(name, "created", callback)` | Installs the limited `created` hook by wrapping the constructor's `connectedCallback`. Other event names are not dispatched by this file. |

Names must satisfy the browser's custom-element rules: lowercase, begin with a
letter, contain a hyphen, and end alphanumeric. Invalid names return `null` or
are assigned the `x-component` fallback by `fromTemplate()`.

### What the helper does not do

- It does not parse component source code or fetch templates.
- It does not know about ATP, VIDI, VICI, or VINI and has no server endpoints.
- It does not sanitize HTML, isolate constructors, or provide a security
  boundary. Registering a constructor executes the application's code with the
  page's privileges.
- `autodefine()` is a presentation convenience (a shadow root containing a
  slot), not a sandbox. Do not use it to make untrusted markup safe.

The `registry`/`handlers` options and the automatic boot behavior are the
actual surface of the current file; older documentation that described page
CRUD, ATP synchronization, or a `window.VENI` API does not apply.

## Optional Go crawler demo

The Go program renders the small form at `/`, calls `/crawl`, and serves its
assets from `/static/`. It uses only the Go standard library.

### Run locally

From this directory:

```sh
gofmt -w main.go
go run .
```

The default listener is `127.0.0.1:8087`. The loopback bind is intentional. The
`/crawl` endpoint is unauthenticated only in this local mode; a non-loopback
listener is refused at startup unless `VENI_API_TOKEN` is set. The token is
checked with a constant-time bearer/`X-Veni-Token` comparison. Open
<http://127.0.0.1:8087/> in a browser.

Configuration is through environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `VENI_HOST` | `127.0.0.1` | Listen address. |
| `VENI_PORT` | `8087` | Listen port. |
| `VENI_API_TOKEN` | unset | Required when binding a non-loopback address; protects `/crawl`. |
| `VENI_ALLOW_PRIVATE_NETWORK` | unset/false | Allows private, loopback, link-local, and other non-public targets. Use only for a trusted local test. |

For example, to crawl a deliberately local fixture while keeping the process
bound to loopback:

```sh
VENI_ALLOW_PRIVATE_NETWORK=true go run .
```

That override removes an important SSRF protection. Never enable it for a
shared or internet-facing process.

### Actual demo routes

| Route | Method | Description |
| --- | --- | --- |
| `/` | `GET` | HTML crawler form. |
| `/crawl?url=<absolute-http(s)-url>&depth=<1..3>&path=<json-array>` | `GET` | Fetches a page and returns a JSON `CrawlNode` tree. `path` is optional and is used for the displayed breadcrumb. |
| `/static/app.js`, `/static/style.css` | `GET` | Demo assets. Other files under `static/` are served by Go's file server. |

Example request:

```sh
curl 'http://127.0.0.1:8087/crawl?url=https%3A%2F%2Fexample.com&depth=1'
```

The response contains `url`, `title`, a short text-only `content` summary,
`depth`, `path`, and child `links` nodes. The browser UI builds the result
tree with DOM nodes and `textContent`; fetched titles, URLs, and text are not
inserted as raw HTML.

### Crawler limits and security boundaries

The demo applies defensive limits, but it is still a developer tool:

- Only `http` and `https` URLs are accepted.
- By default, hostnames and every resolved address are checked to reject
  loopback, private, link-local, unspecified, multicast, and other
  non-global addresses. Redirects are checked again and limited to five hops.
- Requests have a timeout, a bounded response body (1 MiB), a maximum depth
  of 3, and at most 10 links per page.
- HTML is summarized with small regular expressions; this is not a full HTML
  parser and the extracted text is not a security sanitizer.
- A request can trigger at most 128 page visits and 8 MiB of response data,
  with a small process-wide concurrency gate. Request cancellation propagates
  to child requests. These are safety limits, not a general-purpose crawl
  service.
- There is no authorization model, per-user cache, robots.txt handling, or
  durable job queue. The in-memory cache is small and best-effort.
- The browser code treats all crawler values as text, but any page fetched by
  the server is still untrusted input. Do not add a public deployment without
  an allowlist, authentication, egress controls, and an appropriate reverse
  proxy/TLS policy.

### Container image

The Dockerfile uses `golang:1.21-alpine`, matching the `go 1.21` directive in
`go.mod`, and builds the module with `go build .`:

```sh
docker build -t veni-demo .
docker run --rm -p 127.0.0.1:8087:8087 \
  -e VENI_API_TOKEN='replace-with-a-long-random-token' \
  veni-demo
```

The container listens on `0.0.0.0:8087` internally so the published port
works. Set `VENI_API_TOKEN` at runtime; otherwise startup refuses the
non-loopback bind. The example deliberately publishes it only to host loopback.
Add `--network`/egress policy appropriate for the environment if testing
private fixtures; do not make the container internet-accessible by default.

## Checks

From this directory:

```sh
gofmt -w main.go
go test ./...
node --check veni.js
node --check static/app.js
```

`go test ./...` covers the crawler helpers and focused URL/security tests.
There is no npm package, bundler, or browser test suite in this repository.
