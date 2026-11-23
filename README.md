# VENI

VENI has two independent pieces:

1. `veni.js` — a dependency-free browser helper for discovering and registering custom-element tags already present in a page.
2. `main.go` — an optional Go crawler demo with a browser UI.

The browser helper needs no server. The Go demo is a local development tool, not a public crawling service.

Loopback and container-publishing rules for all six demos: [`../docs/LOCAL-DEMOS.md`](../docs/LOCAL-DEMOS.md). Runnable scripts and a loopback-only Compose file: [`../examples`](../examples/README.md).

## Browser helper (`veni.js`)

Exports `window.Veni` (class) and `window.veni` (instance), plus a CommonJS export for tests.

| Method | Purpose |
| --- | --- |
| `init(root)` | Scan the document and `<template>` contents; define tags with registered constructors and mark the rest pending. |
| `define(name, ctor)`, `register(...)`, `defineAll(map)` | Register custom elements through `customElements.define`. Names must be lowercase, hyphenated, and end in a letter or digit. |
| `discover(root)`, `pending()`, `list()`, `isDefined(name)` | Inspect discovered, pending, and defined tags. |
| `autodefine()` | Define lightweight slot-only stubs for pending tags. |
| `fromTemplate(templateEl, ctor)` | Derive a compliant tag name from `data-component`, `data-name`, or a `t-` prefixed template id. |
| `on(name, event, cb)` | Wrap a constructor's `connectedCallback` with a hook. |

`autodefine()` is not a sandbox: its shadow root contains only a slot.

## Go crawler demo

```bash
go test ./...
node --test veni_test.js
go run .                 # http://127.0.0.1:8087
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `VENI_HOST` | `127.0.0.1` | Bind address. |
| `VENI_PORT` | `8087` | Bind port. |
| `VENI_API_TOKEN` | empty | Bearer/`X-Veni-Token` value required on `/crawl` when binding non-loopback. |
| `VENI_ALLOW_PRIVATE_NETWORK` | `false` | Escape hatch for local fixtures. Never enable on a shared deployment. |

Routes:

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Browser UI. |
| `GET /crawl?url=...&depth=...&path=...` | Crawl a public HTTP(S) URL. `depth` defaults to 2 and accepts 1–3; `path` is a JSON array of labels. |
| `GET /static/` | UI assets. |

The crawler validates absolute HTTP(S) URLs, resolves DNS and checks redirects against public-IP rules, and rejects loopback, private, link-local, unspecified, and multicast targets by default. It bounds depth, pages (128), total bytes (8 MiB), per-response bytes (1 MiB), links per page (10), concurrency, and cache size. Results are cached by target, depth, and path. The renderer uses `textContent`, so fetched values are never inserted as markup.

See `SECURITY.md` for the trust model and the limitations of the token guard.

## Tests

`go test ./...` and `node --test veni_test.js` both pass. The Go suite covers the crawler guardrails, URL validation, auth, and HTTP handlers; the JS suite covers discovery, registration, template naming, and lifecycle hooks.

## Remaining gaps

- The token is a single shared demo credential, not per-user authorization. There is no rate limiting, CSRF protection, or TLS.
- The crawler is a guardrail, not a complete SSRF defense. Enforce DNS and network policy at the host or container boundary too.
- `autodefine()` does not make untrusted markup safe, and the browser helper does not sanitize HTML.
- The stale prebuilt binary was removed; build from source with `go build .` (it is now gitignored).
