package veni

import (
	"os"
	"path/filepath"
	"strings"
)

// Registry maps custom element names (e.g., "user-card") to their script URL paths (e.g., "/components/user-card.js").
// It is thread-safe for reads, which is sufficient since it is built once at startup and then read-only during requests.
type Registry struct {
	components map[string]string
}

// NewRegistry creates an empty component registry.
func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]string),
	}
}

// Get retrieves the script path for a given custom element name.
// Returns the path and a boolean indicating if the element was found.
func (r *Registry) Get(name string) (string, bool) {
	path, exists := r.components[name]
	return path, exists
}

// Register adds a component mapping to the registry.
// If the name already exists, it will be overwritten.
func (r *Registry) Register(name, path string) {
	r.components[name] = path
}

// List returns a slice of all registered custom element names.
// Useful for logging or debugging.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

// Discovery handles the scanning of the filesystem to find web component files.
type Discovery struct {
	rootDir string
}

// NewDiscovery creates a new Discovery instance configured to scan the given root directory.
func NewDiscovery(rootDir string) *Discovery {
	return &Discovery{rootDir: rootDir}
}

// Scan walks the directory tree starting at rootDir, finds all .js files,
// extracts the custom element name from each, and returns a populated Registry.
//
// Logic:
// 1. Checks if the directory exists. If not, returns an empty registry.
// 2. Walks the directory recursively.
// 3. Filters for files ending in .js (case-insensitive).
// 4. Reads the file content to look for `customElements.define('name', ...)`.
// 5. If found, uses that name. If not, falls back to the filename (without extension).
// 6. Converts the filesystem path to a URL path relative to the root.
func (d *Discovery) Scan() *Registry {
	registry := NewRegistry()

	// Check if the directory exists
	info, err := os.Stat(d.rootDir)
	if os.IsNotExist(err) {
		// Directory doesn't exist, return empty registry
		return registry
	}
	if err != nil {
		// Some other error occurred (permissions, etc.), log and return empty
		// In a real app, you might want to log this via a logger interface
		return registry
	}
	if !info.IsDir() {
		// The path exists but is not a directory
		return registry
	}

	// Walk the directory tree
	err = filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files/dirs we can't access
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .js files (case-insensitive)
		lowerPath := strings.ToLower(path)
		if !strings.HasSuffix(lowerPath, ".js") {
			return nil
		}

		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files we can't read
			return nil
		}

		// Attempt to extract the component name from the code
		name := ExtractComponentName(content)

		// Fallback: If no define() found, use the filename (without .js)
		if name == "" {
			base := filepath.Base(path)
			name = strings.TrimSuffix(base, filepath.Ext(base))
		}

		// Construct the URL path
		// We want the path relative to the components directory, prefixed with /
		relPath, err := filepath.Rel(d.rootDir, path)
		if err != nil {
			// If we can't get relative path, skip
			return nil
		}

		// Normalize path separators for URL (backslashes to forward slashes)
		urlPath := "/" + filepath.ToSlash(relPath)

		// Register the mapping
		registry.Register(name, urlPath)

		return nil
	})

	if err != nil {
		// If walking failed completely, return what we have (likely empty)
		// In production, you might log this error
	}

	return registry
}

// GetRootDir returns the root directory being scanned.
func (d *Discovery) GetRootDir() string {
	return d.rootDir
}
