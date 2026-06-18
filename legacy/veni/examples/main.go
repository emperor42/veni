package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Emperor42/veni/pkg/veni"
)

func main() {
	// 1. Configuration
	// Define where the web components live and what file extension our templates use.
	cfg := veni.Config{
		ComponentsDir: "./components", // Directory containing .js files
		TemplatesExt:  ".xml",         // Extension for HTML/XML templates
		AutoDiscover:  true,           // Automatically scan for components on startup
	}

	// Optional: Validate configuration
	if err := veni.ValidateConfig(cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// 2. Setup Routes
	mux := http.NewServeMux()

	// Route A: Serve the JavaScript components themselves.
	// We use StripPrefix so that a request to /components/hello-world.js
	// maps to the file ./components/hello-world.js on disk.
	componentsPath := "./components"
	if _, err := os.Stat(componentsPath); os.IsNotExist(err) {
		log.Printf("Warning: Components directory '%s' not found. No components will be injected.", componentsPath)
	} else {
		mux.Handle("/components/", http.StripPrefix("/components/", http.FileServer(http.Dir(componentsPath))))
		log.Printf("Components directory served at /components/")
	}

	// Route B: Serve the HTML/XML templates with VENI middleware.
	// The middleware will intercept requests for .xml files, find custom elements,
	// and inject the necessary <script> tags.
	templatePath := "./templates"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Printf("Warning: Templates directory '%s' not found.", templatePath)
	}

	templateHandler := http.FileServer(http.Dir(templatePath))
	
	// Wrap the FileServer with VENI middleware
	veniedHandler := veni.Middleware(cfg)(templateHandler)
	
	// Handle root path and any other path that isn't /components/
	mux.Handle("/", veniedHandler)

	// 3. Start Server
	port := ":8080"
	
	fmt.Println("--------------------------------------------------")
	fmt.Println("  VENI Example Server")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Templates:   http://localhost%s/\n", port)
	fmt.Printf("  Components:  http://localhost%s/components/\n", port)
	fmt.Printf("  Config:      ComponentsDir=%s, TemplatesExt=%s\n", cfg.ComponentsDir, cfg.TemplatesExt)
	fmt.Println("--------------------------------------------------")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("--------------------------------------------------")

	log.Printf("Starting server on http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// init ensures the example directories exist for a smooth first run experience.
// In a real production app, you would likely remove this and assume directories exist.
func init() {
	dirs := []string{"./templates", "./components"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Could not create directory %s: %v", dir, err)
			}
		}
	}

	// Create a sample template if it doesn't exist
	templateFile := filepath.Join("./templates", "index.xml")
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		content := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>VENI Demo</title>
    <style>body { font-family: sans-serif; padding: 2rem; }</style>
</head>
<body>
    <h1>VENI Web Components Demo</h1>
    <p>If you see colored boxes below, VENI is working!</p>
    
    <!-- These tags will be auto-detected and scripts injected -->
    <hello-world name="User"></hello-world>
    <user-card name="Alice" email="alice@example.com"></user-card>
    <stats-panel total="100" active="25"></stats-panel>
</body>
</html>`
		if err := os.WriteFile(templateFile, []byte(content), 0644); err != nil {
			log.Printf("Could not create sample template: %v", err)
		} else {
			log.Printf("Created sample template: %s", templateFile)
		}
	}

	// Create sample components if they don't exist
	samples := map[string]string{
		"./components/hello-world.js": `class HelloWorld extends HTMLElement {
    constructor() { super(); this.attachShadow({mode: 'open'}); }
    connectedCallback() {
        const name = this.getAttribute('name') || 'Guest';
        this.shadowRoot.innerHTML = \`
            <style>:host { display: block; padding: 1rem; background: #667eea; color: white; border-radius: 8px; margin: 1rem 0; }</style>
            <h2>Hello, \${name}!</h2>
        \`;
    }
}
customElements.define('hello-world', HelloWorld);`,
		"./components/user-card.js": `class UserCard extends HTMLElement {
    constructor() { super(); this.attachShadow({mode: 'open'}); }
    connectedCallback() {
        const name = this.getAttribute('name') || 'Unknown';
        const email = this.getAttribute('email') || '';
        this.shadowRoot.innerHTML = \`
            <style>:host { display: block; border: 1px solid #ddd; padding: 1rem; margin: 1rem 0; border-radius: 8px; background: white; }</style>
            <strong>\$${name}</strong><br><small>\$${email}</small>
        \`;
    }
}
customElements.define('user-card', UserCard);`,
		"./components/stats-panel.js": `class StatsPanel extends HTMLElement {
    constructor() { super(); this.attachShadow({mode: 'open'}); }
    connectedCallback() {
        const total = this.getAttribute('total') || '0';
        const active = this.getAttribute('active') || '0';
        this.shadowRoot.innerHTML = \`
            <style>:host { display: block; text-align: center; padding: 1rem; background: #f0f0f0; margin: 1rem 0; border-radius: 8px; }</style>
            <div>Total: <strong>\${total}</strong></div>
            <div>Active: <strong>\${active}</strong></div>
        \`;
    }
}
customElements.define('stats-panel', StatsPanel);`,
	}

	for path, content := range samples {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				log.Printf("Could not create sample component %s: %v", path, err)
			} else {
				log.Printf("Created sample component: %s", path)
			}
		}
	}
}