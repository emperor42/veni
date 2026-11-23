package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CrawlNode struct {
	URL     string       `json:"url"`
	Title   string       `json:"title"`
	Content string       `json:"content"`
	Depth   int          `json:"depth"`
	Path    []string     `json:"path"`
	Links   []*CrawlNode `json:"links"`
}

const (
	defaultHost            = "127.0.0.1"
	defaultPort            = "8087"
	maxCrawlDepth          = 3
	maxBodyBytes     int64 = 1 << 20
	maxLinksPerPage        = 10
	maxCrawlPages          = 128
	maxCrawlBytes    int64 = 8 << 20
	crawlTimeout           = 12 * time.Second
	crawlCacheLimit        = 64
	crawlConcurrency       = 4
)

var (
	crawlCache = make(map[string]*CrawlNode)
	crawlSlots = make(chan struct{}, crawlConcurrency)
	cacheMu    sync.RWMutex
	linkRe     = regexp.MustCompile(`<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>(.*?)</a>`)
	titleRe    = regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	bodyRe     = regexp.MustCompile(`<body[^>]*>(.*?)</body>`)
	stripRe    = regexp.MustCompile(`<[^>]+>`)
	maxDepth   = 2
)

type crawler struct {
	client       *http.Client
	allowPrivate bool
	ctx          context.Context

	maxPages  int
	pages     int
	maxBytes  int64
	bytesRead int64

	mu      sync.Mutex
	visited map[string]struct{}
}

func main() {
	addr := listenAddress()
	if err := validateListenSecurity(addr); err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("veni crawler demo listening on http://%s (public HTTP(S) targets only)", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.Handle("/crawl", crawlAuth(http.HandlerFunc(crawlHandler)))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return demoSecurityHeaders(mux)
}

func crawlToken() string { return strings.TrimSpace(os.Getenv("VENI_API_TOKEN")) }

func crawlAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := crawlToken()
		if want == "" {
			// An unconfigured listener is intentionally a loopback-only local
			// demo. main refuses a non-loopback bind without a token.
			next.ServeHTTP(w, r)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Veni-Token"))
		if auth := strings.TrimSpace(r.Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			got = strings.TrimSpace(auth[len("Bearer "):])
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="veni"`)
			http.Error(w, "crawler authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func demoSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func validateListenSecurity(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", addr, err)
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	if crawlToken() == "" {
		return fmt.Errorf("refusing non-loopback listen address %q without VENI_API_TOKEN", addr)
	}
	return nil
}

func listenAddress() string {
	host := strings.TrimSpace(os.Getenv("VENI_HOST"))
	if host == "" {
		host = defaultHost
	}
	port := strings.TrimSpace(os.Getenv("VENI_PORT"))
	if port == "" {
		port = defaultPort
	}
	return net.JoinHostPort(host, port)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "template unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("veni: render home page: %v", err)
	}
}

func crawlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	select {
	case crawlSlots <- struct{}{}:
		defer func() { <-crawlSlots }()
	default:
		http.Error(w, "crawler is busy; retry shortly", http.StatusTooManyRequests)
		return
	}

	target := strings.TrimSpace(r.URL.Query().Get("url"))
	if target == "" {
		http.Error(w, "url parameter required", http.StatusBadRequest)
		return
	}

	depth := maxDepth
	if depthText := strings.TrimSpace(r.URL.Query().Get("depth")); depthText != "" {
		parsedDepth, err := strconv.Atoi(depthText)
		if err != nil || parsedDepth < 1 || parsedDepth > maxCrawlDepth {
			http.Error(w, fmt.Sprintf("depth must be between 1 and %d", maxCrawlDepth), http.StatusBadRequest)
			return
		}
		depth = parsedDepth
	}

	var path []string
	if pathText := strings.TrimSpace(r.URL.Query().Get("path")); pathText != "" {
		if err := json.Unmarshal([]byte(pathText), &path); err != nil || len(path) > maxCrawlDepth+1 {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		for _, entry := range path {
			if strings.TrimSpace(entry) == "" || len(entry) > 2048 {
				http.Error(w, "invalid path", http.StatusBadRequest)
				return
			}
		}
	}

	allowPrivate := allowPrivateNetwork()
	crawlContext, cancel := context.WithTimeout(r.Context(), crawlTimeout)
	defer cancel()
	parsedTarget, err := validateCrawlURL(crawlContext, target, allowPrivate)
	if err != nil {
		http.Error(w, "url is not an allowed HTTP(S) target", http.StatusBadRequest)
		return
	}
	target = parsedTarget.String()
	pathKey, _ := json.Marshal(path)
	cacheKey := target + "\x00" + strconv.Itoa(depth) + "\x00" + string(pathKey)

	cacheMu.RLock()
	cached, ok := crawlCache[cacheKey]
	cacheMu.RUnlock()
	if ok {
		writeJSON(w, cached)
		return
	}

	crawler := newCrawler(allowPrivate)
	crawler.ctx = crawlContext
	crawler.maxPages = maxCrawlPages
	crawler.maxBytes = maxCrawlBytes
	node := crawler.crawl(target, depth, path)
	if node != nil {
		cacheMu.Lock()
		if len(crawlCache) >= crawlCacheLimit {
			// The demo cache is intentionally small and best-effort. Remove an
			// arbitrary entry rather than allowing unbounded memory growth.
			for key := range crawlCache {
				delete(crawlCache, key)
				break
			}
		}
		crawlCache[cacheKey] = node
		cacheMu.Unlock()
	}
	writeJSON(w, node)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("veni: encode crawl response: %v", err)
	}
}

func allowPrivateNetwork() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("VENI_ALLOW_PRIVATE_NETWORK")))
	return value == "1" || value == "true" || value == "yes"
}

func newCrawler(allowPrivate bool) *crawler {
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	c := &crawler{
		allowPrivate: allowPrivate,
		ctx:          context.Background(),
		maxPages:     maxCrawlPages,
		maxBytes:     maxCrawlBytes,
		visited:      make(map[string]struct{}),
	}
	transport := &http.Transport{
		Proxy: nil, // Do not silently route public targets through an environment proxy.
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialCrawlAddress(ctx, dialer, network, address, allowPrivate)
		},
		MaxIdleConns:          16,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	c.client = &http.Client{
		Transport: transport,
		Timeout:   crawlTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			_, err := validateCrawlURL(req.Context(), req.URL.String(), allowPrivate)
			return err
		},
	}
	return c
}

func dialCrawlAddress(ctx context.Context, dialer *net.Dialer, network, address string, allowPrivate bool) (net.Conn, error) {
	if allowPrivate {
		return dialer.DialContext(ctx, network, address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid crawl address")
	}
	ips, err := resolveHost(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses for crawl host")
	}

	var lastErr error
	for _, ip := range ips {
		if isBlockedIP(ip) {
			continue
		}
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("all addresses for crawl host are blocked")
}

func resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" {
		return nil, fmt.Errorf("empty crawl host")
	}
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	return ips, nil
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || !ip.IsGlobalUnicast()
}

func validateCrawlURL(ctx context.Context, raw string, allowPrivate bool) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\t") {
		return nil, fmt.Errorf("invalid crawl URL")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid crawl URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported crawl scheme")
	}
	if u.User != nil || u.Hostname() == "" || strings.Contains(u.Hostname(), "%") {
		return nil, fmt.Errorf("invalid crawl host")
	}
	if port := u.Port(); port != "" {
		parsedPort, err := strconv.Atoi(port)
		if err != nil || parsedPort < 1 || parsedPort > 65535 {
			return nil, fmt.Errorf("invalid crawl port")
		}
	}
	u.Fragment = ""
	if !allowPrivate {
		host := strings.ToLower(u.Hostname())
		if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
			return nil, fmt.Errorf("private crawl host is disabled")
		}
		ips, err := resolveHost(ctx, host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("could not resolve crawl host")
		}
		for _, ip := range ips {
			if isBlockedIP(ip) {
				return nil, fmt.Errorf("private crawl address is disabled")
			}
		}
	}
	return u, nil
}

func crawl(target string, depth int, path []string) *CrawlNode {
	ctx, cancel := context.WithTimeout(context.Background(), crawlTimeout)
	defer cancel()
	c := newCrawler(allowPrivateNetwork())
	c.ctx = ctx
	c.maxPages = maxCrawlPages
	c.maxBytes = maxCrawlBytes
	return c.crawl(target, depth, path)
}

func (c *crawler) crawl(target string, depth int, path []string) *CrawlNode {
	if depth <= 0 {
		return nil
	}
	if c.maxPages <= 0 {
		c.maxPages = maxCrawlPages
	}
	if c.maxBytes <= 0 {
		c.maxBytes = maxCrawlBytes
	}
	if c.pages >= c.maxPages || c.bytesRead >= c.maxBytes {
		return nil
	}
	c.pages++

	parentContext := c.ctx
	if parentContext == nil {
		parentContext = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentContext, crawlTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil
	}
	parsed, err := validateCrawlURL(ctx, target, c.allowPrivate)
	if err != nil {
		return crawlErrorNode(target, "URL is not an allowed HTTP(S) target", depth, path)
	}
	target = parsed.String()

	c.mu.Lock()
	if _, seen := c.visited[target]; seen {
		c.mu.Unlock()
		return nil
	}
	c.visited[target] = struct{}{}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return crawlErrorNode(target, "Could not create crawl request", depth, path)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "veni-local-demo/1.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return crawlErrorNode(target, "Target could not be fetched", depth, path)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return crawlErrorNode(target, fmt.Sprintf("Target returned HTTP %d", resp.StatusCode), depth, path)
	}

	remaining := c.maxBytes - c.bytesRead
	if remaining <= 0 {
		return nil
	}
	readLimit := maxBodyBytes
	if remaining < readLimit {
		readLimit = remaining
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, readLimit+1))
	if readErr != nil {
		return crawlErrorNode(target, "Target response could not be read", depth, path)
	}
	if int64(len(body)) > readLimit {
		// A page larger than the per-request limit is truncated as before; if
		// it would exceed the whole-crawl budget, stop expanding the tree.
		if readLimit == remaining {
			return nil
		}
		body = body[:readLimit]
	}
	c.bytesRead += int64(len(body))
	page := string(body)
	newPath := appendPath(path, target)
	node := &CrawlNode{
		URL:     target,
		Title:   extractTitle(page),
		Content: extractContent(page),
		Depth:   depth,
		Path:    newPath,
	}

	links := extractLinks(page, target)
	childDepth := depth - 1
	if childDepth <= 0 {
		return node
	}
	for _, link := range links {
		child := c.crawl(link, childDepth, newPath)
		if child != nil {
			node.Links = append(node.Links, child)
		}
	}
	return node
}

func crawlErrorNode(target, message string, depth int, path []string) *CrawlNode {
	return &CrawlNode{
		URL:   target,
		Title: "Error: " + message,
		Depth: depth,
		Path:  appendPath(path, target),
	}
}

func appendPath(path []string, target string) []string {
	out := make([]string, 0, len(path)+1)
	for _, entry := range path {
		if strings.TrimSpace(entry) != "" && len(entry) <= 2048 {
			out = append(out, entry)
		}
	}
	return append(out, target)
}

func extractTitle(page string) string {
	m := titleRe.FindStringSubmatch(page)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return "No Title"
}

func extractContent(page string) string {
	m := bodyRe.FindStringSubmatch(page)
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

func extractLinks(page, baseURL string) []string {
	matches := linkRe.FindAllStringSubmatch(page, -1)
	seen := make(map[string]bool)
	var links []string
	for _, match := range matches {
		link := strings.TrimSpace(match[1])
		if link == "" || strings.HasPrefix(link, "#") {
			continue
		}
		abs := resolveURL(link, baseURL)
		if abs != "" && !seen[abs] {
			seen[abs] = true
			links = append(links, abs)
		}
		if len(links) == maxLinksPerPage {
			break
		}
	}
	return links
}

func resolveURL(href, base string) string {
	href = strings.TrimSpace(href)
	if href == "" || len(href) > 2048 || strings.ContainsAny(href, "\\\r\n\t") {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" || baseURL.User != nil {
		return ""
	}
	hrefURL, err := url.Parse(href)
	if err != nil || (hrefURL.Scheme != "" && hrefURL.Scheme != "http" && hrefURL.Scheme != "https") {
		return ""
	}
	resolved := baseURL.ResolveReference(hrefURL)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	resolved.Fragment = ""
	return resolved.String()
}
