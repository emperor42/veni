package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type CrawlNode struct {
	URL     string      `json:"url"`
	Title   string      `json:"title"`
	Content string      `json:"content"`
	Depth   int         `json:"depth"`
	Path    []string    `json:"path"`
	Links   []*CrawlNode `json:"links"`
}

var (
	crawlCache = make(map[string]*CrawlNode)
	cacheMu    sync.RWMutex
	linkRe     = regexp.MustCompile(`<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>(.*?)</a>`)
	titleRe    = regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	bodyRe     = regexp.MustCompile(`<body[^>]*>(.*?)</body>`)
	stripRe    = regexp.MustCompile(`<[^>]+>`)
	maxDepth   = 2
)

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/crawl", crawlHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8085", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func crawlHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		http.Error(w, "url parameter required", http.StatusBadRequest)
		return
	}
	depthStr := r.URL.Query().Get("depth")
	depth := maxDepth
	if depthStr != "" {
		fmt.Sscanf(depthStr, "%d", &depth)
	}
	pathStr := r.URL.Query().Get("path")
	var path []string
	if pathStr != "" {
		json.Unmarshal([]byte(pathStr), &path)
	}

	cacheMu.RLock()
	if node, ok := crawlCache[target]; ok {
		cacheMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(node)
		return
	}
	cacheMu.RUnlock()

	node := crawl(target, depth, path)
	if node != nil {
		cacheMu.Lock()
		crawlCache[target] = node
		cacheMu.Unlock()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

func crawl(target string, depth int, path []string) *CrawlNode {
	if depth <= 0 {
		return nil
	}

	resp, err := http.Get(target)
	if err != nil {
		return &CrawlNode{URL: target, Title: "Error: " + err.Error(), Depth: depth, Path: path}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	title := extractTitle(html)
	content := extractContent(html)
	newPath := append(append([]string{}, path...), target)

	node := &CrawlNode{
		URL:     target,
		Title:   title,
		Content: content,
		Depth:   depth,
		Path:    newPath,
	}

	links := extractLinks(html, target)
	childDepth := depth - 1
	if childDepth > 0 {
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, l := range links {
			wg.Add(1)
			go func(linkURL string) {
				defer wg.Done()
				child := crawl(linkURL, childDepth, newPath)
				if child != nil {
					mu.Lock()
					node.Links = append(node.Links, child)
					mu.Unlock()
				}
			}(l)
		}
		wg.Wait()
	}

	return node
}

func extractTitle(html string) string {
	m := titleRe.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return "No Title"
}

func extractContent(html string) string {
	m := bodyRe.FindStringSubmatch(html)
	if len(m) > 1 {
		text := stripRe.ReplaceAllString(m[1], " ")
		text = strings.Join(strings.Fields(text), " ")
		if len(text) > 500 {
			text = text[:500] + "..."
		}
		return text
	}
	return "No content"
}

func extractLinks(html, baseURL string) []string {
	matches := linkRe.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		link := m[1]
		link = strings.TrimSpace(link)
		if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
			continue
		}
		abs := resolveURL(link, baseURL)
		if abs != "" && !seen[abs] {
			seen[abs] = true
			links = append(links, abs)
		}
	}
	if len(links) > 10 {
		links = links[:10]
	}
	return links
}

func resolveURL(href, base string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	hrefURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(hrefURL).String()
}
