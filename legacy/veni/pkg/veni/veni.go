package veni

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

// Config holds the configuration for the VENI middleware.
type Config struct {
	// ComponentsDir is the directory containing web component JS files.
	// This path is resolved relative to the current working directory
	// unless it starts with a slash (absolute path).
	ComponentsDir string

	// TemplatesExt is the file extension for template files to process.
	// Commonly ".xml" or ".html". Defaults to ".xml" if empty.
	TemplatesExt string

	// AutoDiscover enables automatic component discovery on startup.
	// If true, the middleware scans ComponentsDir once when initialized.
	// If false, the registry must be populated manually or via a custom
	// discovery mechanism before the middleware is used.
	AutoDiscover bool
}

// Middleware returns an http.Handler that wraps the provided handler.
// It intercepts responses for template files, scans for custom elements,
// and injects the necessary <script> tags to load the corresponding components.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	// Set defaults
	if cfg.TemplatesExt == "" {
		cfg.TemplatesExt = ".xml"
	}
	if cfg.ComponentsDir == "" {
		cfg.ComponentsDir = "components"
	}

	// Initialize the component registry
	var registry *Registry
	if cfg.AutoDiscover {
		discovery := NewDiscovery(cfg.ComponentsDir)
		registry = discovery.Scan()
		if len(registry.List()) > 0 {
			fmt.Printf("[VENI] Discovered %d components in %s\n", len(registry.List()), cfg.ComponentsDir)
		} else {
			fmt.Printf("[VENI] Warning: No components found in %s\n", cfg.ComponentsDir)
		}
	} else {
		registry = NewRegistry()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Optimization: Only process requests ending with the template extension
			if !strings.HasSuffix(r.URL.Path, cfg.TemplatesExt) {
				next.ServeHTTP(w, r)
				return
			}

			// Capture the response to inspect and modify the body
			rec := NewResponseRecorder(w)

			// Call the next handler (usually http.FileServer)
			next.ServeHTTP(rec, r)

			// If the response was an error (4xx, 5xx) or redirect, pass it through unchanged
			if rec.Status() >= 400 {
				rec.Flush()
				return
			}

			// Get the captured body
			body := rec.Body()
			if len(body) == 0 {
				rec.Flush()
				return
			}

			// Process the template: find elements and inject scripts
			processedBody, err := ProcessTemplate(body, registry, cfg.ComponentsDir)
			if err != nil {
				// Log error but serve original content to prevent breaking the page
				fmt.Printf("[VENI] Error processing template %s: %v\n", r.URL.Path, err)
				rec.Flush()
				return
			}

			// If no changes were made, flush original
			if bytes.Equal(body, processedBody) {
				rec.Flush()
				return
			}

			// Update headers for the modified content
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Del("Content-Length")           // Remove old length, new length is unknown until write
			w.Header().Set("X-VENI-Processed", "true") // Debug header

			// Write the modified content
			w.Write(processedBody)
		})
	}
}

// FileServer is a convenience function that creates a standard http.FileServer
// wrapped with the VENI middleware.
//
// Usage:
//
//	cfg := veni.Config{ComponentsDir: "./components", TemplatesExt: ".xml"}
//	fs := veni.FileServer(http.Dir("./templates"), cfg)
//	http.Handle("/", fs)
func FileServer(root http.Dir, cfg Config) http.Handler {
	handler := http.FileServer(root)
	return Middleware(cfg)(handler)
}

// ProcessTemplate is the core logic that transforms raw HTML/XML content.
// It identifies custom elements, looks up their script paths in the registry,
// and injects the necessary <script> tags.
func ProcessTemplate(html []byte, registry *Registry, componentsDir string) ([]byte, error) {
	if registry == nil {
		return html, nil
	}

	parser := NewParser(html)

	// Step 1: Find all custom element usages (e.g., <my-component>)
	elements := parser.FindCustomElements()
	if len(elements) == 0 {
		return html, nil
	}

	// Step 2: Determine which scripts need to be injected
	var scriptsToInject []string
	for _, elem := range elements {
		if path, exists := registry.Get(elem); exists {
			// Ensure the path is relative to the web root (starts with /)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			scriptsToInject = append(scriptsToInject, path)
		}
	}

	if len(scriptsToInject) == 0 {
		return html, nil
	}

	// Step 3: Inject the scripts
	// We remove duplicates just in case the same element appears multiple times
	uniqueScripts := make([]string, 0, len(scriptsToInject))
	seen := make(map[string]bool)
	for _, s := range scriptsToInject {
		if !seen[s] {
			seen[s] = true
			uniqueScripts = append(uniqueScripts, s)
		}
	}

	return parser.InjectScripts(uniqueScripts, componentsDir), nil
}

// ValidateConfig performs basic validation on the configuration.
// It returns an error if the configuration is invalid.
func ValidateConfig(cfg Config) error {
	if cfg.ComponentsDir == "" {
		return fmt.Errorf("ComponentsDir cannot be empty")
	}
	if cfg.TemplatesExt == "" {
		return fmt.Errorf("TemplatesExt cannot be empty")
	}
	// Ensure extension starts with a dot
	if !strings.HasPrefix(cfg.TemplatesExt, ".") {
		cfg.TemplatesExt = "." + cfg.TemplatesExt
	}
	return nil
}
