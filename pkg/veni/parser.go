package veni

import (
	"bytes"
	"regexp"
	"strings"
)

// Parser handles HTML/XML parsing and modification for the VENI middleware.
// It is responsible for finding custom element tags and injecting script references.
type Parser struct {
	content []byte
}

// NewParser creates a new Parser instance with the provided HTML content.
func NewParser(content []byte) *Parser {
	return &Parser{content: content}
}

// customElementRegex matches custom element tags.
// Pattern breakdown:
//   - < : Opening bracket
//   - ([a-z][a-z0-9]*-[a-z0-9-]*) : Capture group for the tag name.
//   - Must start with a lowercase letter.
//   - Must contain at least one hyphen (required by Web Components spec).
//   - Can contain letters, numbers, and hyphens.
//   - [^>]* : Any characters except '>' (attributes, spaces, etc.)
//   - > : Closing bracket
//
// This regex ignores closing tags (e.g., </my-element>) because they don't start with '<' followed by a name.
var customElementRegex = regexp.MustCompile(`<([a-z][a-z0-9]*-[a-z0-9-]*)[^>]*>`)

// FindCustomElements extracts all unique custom element names from the HTML content.
// It returns a slice of strings representing the tag names found (e.g., ["user-card", "hello-world"]).
func (p *Parser) FindCustomElements() []string {
	matches := customElementRegex.FindAllSubmatch(p.content, -1)

	seen := make(map[string]bool)
	var elements []string

	for _, match := range matches {
		if len(match) > 1 {
			name := string(match[1])
			if !seen[name] {
				seen[name] = true
				elements = append(elements, name)
			}
		}
	}

	return elements
}

// InjectScripts generates <script> tags for the provided component paths
// and injects them into the HTML content.
//
// Injection Strategy:
// 1. Attempts to insert before the closing </head> tag.
// 2. If not found, attempts to insert before the closing </body> tag.
// 3. If neither is found, appends to the very end of the document.
//
// The 'baseDir' argument is currently unused in the injection logic but
// is kept for potential future path resolution or logging.
func (p *Parser) InjectScripts(paths []string, baseDir string) []byte {
	if len(paths) == 0 {
		return p.content
	}

	// Build the script tags string
	var scriptTags bytes.Buffer
	scriptTags.WriteString("\n<!-- VENI: Auto-injected web components -->\n")

	for _, path := range paths {
		// Ensure path starts with / for consistent URL handling
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		scriptTags.WriteString(`<script src="`)
		scriptTags.WriteString(path)
		scriptTags.WriteString(`"></script>`)
		scriptTags.WriteString("\n")
	}

	injectionPoint := scriptTags.Bytes()

	// Strategy 1: Inject before </head>
	headClose := bytes.LastIndex(p.content, []byte("</head>"))
	if headClose != -1 {
		result := make([]byte, 0, len(p.content)+len(injectionPoint))
		result = append(result, p.content[:headClose]...)
		result = append(result, injectionPoint...)
		result = append(result, p.content[headClose:]...)
		return result
	}

	// Strategy 2: Inject before </body>
	bodyClose := bytes.LastIndex(p.content, []byte("</body>"))
	if bodyClose != -1 {
		result := make([]byte, 0, len(p.content)+len(injectionPoint))
		result = append(result, p.content[:bodyClose]...)
		result = append(result, injectionPoint...)
		result = append(result, p.content[bodyClose:]...)
		return result
	}

	// Strategy 3: Append to end (fallback)
	result := make([]byte, 0, len(p.content)+len(injectionPoint))
	result = append(result, p.content...)
	result = append(result, injectionPoint...)
	return result
}

// ExtractComponentName parses a JavaScript file's content to find the custom element name it defines.
// It looks for the pattern: customElements.define('element-name', ...)
// If the pattern is not found, it returns an empty string.
var defineRegex = regexp.MustCompile(`customElements\s*\.\s*define\s*\(\s*['"]([a-z][a-z0-9]*-[a-z0-9-]*)['"]`)

// ExtractComponentName returns the custom element name found in the JS content.
func ExtractComponentName(jsContent []byte) string {
	match := defineRegex.FindSubmatch(jsContent)
	if len(match) > 1 {
		return string(match[1])
	}
	return ""
}

// IsCustomElementTag checks if a given tag name is a valid custom element.
// A valid custom element name must contain a hyphen.
func IsCustomElementTag(tagName string) bool {
	return strings.Contains(tagName, "-")
}
