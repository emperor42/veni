# VENI security notes

## Scope

VENI has a browser helper (`veni.js`) and an optional Go crawler demo. They are
separate trust surfaces. The browser helper has no server-side security model;
the crawler is a deliberately small local development tool. It is loopback-only
by default; a non-loopback bind requires `VENI_API_TOKEN`, which is checked on
`/crawl`, but this is not a user/authorization system.

## Browser helper

`veni.js` only uses DOM and Custom Elements APIs. It does not fetch remote
code, sanitize HTML, or provide an isolation boundary. A constructor passed to
`define()` or `fromTemplate()` runs with the same privileges as the page. The
`autodefine()` stub uses an open shadow root containing a slot, which is not a
sandbox and does not make untrusted content safe.

Applications remain responsible for controlling the scripts they load, using a
trusted Content Security Policy, and treating any constructor or template data
as code rather than user text.

## Go crawler demo

The demo binds to `127.0.0.1:8087` by default and accepts only `GET` requests
for its UI/crawl routes. Its crawler has the following safeguards:

- absolute `http` and `https` URLs only;
- public-IP validation by default, including DNS resolution and redirect
  checks;
- loopback, private, link-local, unspecified, multicast, and other
  non-global addresses rejected;
- request cancellation propagation, one overall crawl deadline, bounded pages
  and bytes, bounded depth/link fan-out, a process-wide concurrency gate, and
  a bounded in-memory cache.

These are guardrails, not a complete SSRF defense. DNS and network policy
should still be enforced at the host/container boundary. The
`VENI_ALLOW_PRIVATE_NETWORK=true` escape hatch is intended only for trusted
local fixtures; it permits requests to services such as local databases and
cloud metadata addresses and must never be enabled on a shared deployment.

The crawler has no user accounts, per-user rate limit, CSRF protection, TLS,
reverse proxy, or public-service hardening. Do not publish it as an open proxy.
The Dockerfile exposes the process internally on port 8087; provide
`VENI_API_TOKEN` and publish that port only to a protected network unless a
separately reviewed policy is in place.

The browser renderer uses DOM `textContent` and does not insert crawler titles,
URLs, or summaries as raw HTML. This prevents the demo UI from treating a
fetched value as markup, but it does not make fetched content trustworthy.

## Reporting a suspected vulnerability

Report suspected vulnerabilities privately to the project maintainers rather
than placing sensitive targets, credentials, or exploit details in a public
issue. Include the affected file, a minimal reproduction, and the deployment
assumptions (especially whether private-network access was enabled).
