# VENI

**Veni** (Latin for "I come") is a lightweight, zero-dependency Go middleware designed to automatically discover and inject vanilla JavaScript web components into HTML templates served via the standard `net/http` FileServer.

Built strictly with the Go standard library, VENI bridges the gap between static XML/HTML templates and dynamic web components without requiring a build step or external templating engine.

## Features

- **Zero Dependencies**: Uses only `net/http`, `os`, `path/filepath`, and `regexp`.
- **Automatic Discovery**: Scans a designated directory for `.js` files and registers custom elements.
- **Smart Injection**: Detects custom element usage in templates and injects the necessary `<script>` tags before `</head>` or `</body>`.
- **Composable**: Designed to be layered with other middlewares (e.g., VIDI for data, VICI for context).
- **Standard Library Only**: No external imports.

## Installation

```bash
go get github.com/Emperor42/veni