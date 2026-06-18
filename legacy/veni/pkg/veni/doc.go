/*
Package veni provides a lightweight, zero-dependency middleware for Go's standard
library net/http package. It automatically discovers vanilla JavaScript web components
and injects the necessary <script> tags into HTML/XML templates served by a FileServer.

VENI (part of the Veni, Vidi, Vici project suite) bridges the gap between static
server-side templates and dynamic client-side web components without requiring
a build step, bundler, or external templating engine.

# Key Features

  - Zero External Dependencies: Built entirely with the Go standard library.
  - Automatic Discovery: Scans a designated directory for .js files and registers
    custom elements defined within them.
  - Smart Injection: Detects custom element usage in templates and injects
    corresponding <script> tags before </head> or </body>.
  - Composable: Designed to be layered with other middlewares (e.g., VIDI for data,
    VICI for context) or used standalone.
  - Standard Library Only: No imports outside of "net/http", "os", "path/filepath",
    "regexp", and "bytes".

# How It Works

 1. Discovery: When the middleware is initialized, it scans the configured
    ComponentsDir for .js files. It parses each file to find
    customElements.define('tag-name', ...) calls. If found, it maps 'tag-name'
    to the file's URL path. If not found, it falls back to the filename.

 2. Interception: When a request for a template file (e.g., .xml) arrives,
    VENI captures the response from the underlying handler (usually http.FileServer).

 3. Parsing: It scans the HTML content for tags matching the custom element pattern
    (e.g., <my-component>).

 4. Injection: It injects <script src="/components/my-component.js"></script> tags
    into the HTML, prioritizing insertion before </head>.

 5. Response: The modified HTML is sent to the client, allowing the browser to
    load and render the components automatically.

# Usage

Basic Setup:

	import (
	    "log"
	    "net/http"
	    "github.com/Emperor42/veni/pkg/veni"
	)

	func main() {
	    cfg := veni.Config{
	        ComponentsDir: "./components",
	        TemplatesExt:  ".xml",
	        AutoDiscover:  true,
	    }

	    // Serve templates with VENI middleware
	    fs := http.FileServer(http.Dir("./templates"))
	    handler := veni.Middleware(cfg)(fs)

	    // Serve components themselves
	    http.Handle("/components/", http.StripPrefix("/components/", http.FileServer(http.Dir("./components"))))

	    http.Handle("/", handler)

	    log.Println("Server starting on :8080")
	    log.Fatal(http.ListenAndServe(":8080", nil))
	}

Convenience Function:

	// Equivalent to the manual setup above
	cfg := veni.Config{ComponentsDir: "./components"}
	http.Handle("/", veni.FileServer(http.Dir("./templates"), cfg))

# Configuration

The veni.Config struct allows customization:

  - ComponentsDir: Directory containing web component JS files (default: "components").
  - TemplatesExt: File extension for templates to process (default: ".xml").
  - AutoDiscover: Whether to scan the directory on startup (default: true).

# Custom Element Naming

VENI expects custom elements to follow the standard Web Components specification:
tag names must contain a hyphen (e.g., <user-card>, <hello-world>).
Standard HTML tags (div, span, etc.) are ignored.

The system identifies components in two ways:
1. By parsing the JS file for customElements.define('name', ...).
2. By using the filename if no define() call is found (e.g., my-widget.js -> my-widget).

# Limitations

  - Regex-based Parsing: Uses regular expressions for HTML parsing. While robust
    for well-formed templates, it may not handle malformed HTML perfectly.
  - Single-Pass: Components are discovered once at startup. Dynamic addition of
    components requires restarting the server (unless a watch mechanism is added).
  - No Transpilation: Does not transpile modern JS; browsers must support ES6 classes.

# Integration with VIDI and VICI

VENI is designed to be the first layer in a stack:
  - VENI: Handles UI rendering and component injection.
  - VIDI: Handles form data conversion and storage (SQL/NoSQL/FS).
  - VICI: Handles session context and user tracking.

These middlewares can be chained:

	handler := veni.Middleware(veniCfg)(
	    vidi.Middleware(vidiCfg)(
	        vici.Middleware(viciCfg)(http.FileServer(http.Dir(".")))
	    )
	)

# License

MIT License. See LICENSE file for details.
*/
package veni
