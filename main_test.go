package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateCrawlURLRejectsPrivateAndNonHTTP(t *testing.T) {
	ctx := context.Background()
	for _, target := range []string{
		"http://127.0.0.1:8080/private",
		"http://[::1]/private",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
		"ftp://example.com/file",
		"http://user:password@example.com/",
	} {
		if _, err := validateCrawlURL(ctx, target, false); err == nil {
			t.Errorf("validateCrawlURL(%q) allowed an unsafe target", target)
		}
	}
}

func TestValidateCrawlURLPrivateOverride(t *testing.T) {
	parsed, err := validateCrawlURL(context.Background(), "http://127.0.0.1:8080/", true)
	if err != nil {
		t.Fatalf("private override rejected target: %v", err)
	}
	if parsed.Hostname() != "127.0.0.1" {
		t.Fatalf("parsed host = %q, want 127.0.0.1", parsed.Hostname())
	}
}

func TestResolveURLRejectsUnsafeSchemes(t *testing.T) {
	base := "https://example.com/docs/index.html"
	if got := resolveURL("../next", base); got != "https://example.com/next" {
		t.Fatalf("relative URL = %q", got)
	}
	for _, href := range []string{"javascript:alert(1)", "data:text/html,hello", "file:///etc/passwd"} {
		if got := resolveURL(href, base); got != "" {
			t.Errorf("resolveURL(%q) = %q, want empty", href, got)
		}
	}
}

func TestCrawlerFetchesFixture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Fixture &amp; page</title></head><body><p>Hello <strong>crawler</strong></p></body></html>`))
	}))
	defer server.Close()

	node := newCrawler(true).crawl(server.URL, 1, nil)
	if node == nil {
		t.Fatal("crawl returned nil")
	}
	if node.Title != "Fixture &amp; page" {
		t.Errorf("title = %q", node.Title)
	}
	if !strings.Contains(node.Content, "Hello crawler") {
		t.Errorf("content = %q", node.Content)
	}
	if len(node.Path) != 1 || node.Path[0] != server.URL {
		t.Errorf("path = %v", node.Path)
	}
}

func TestCrawlerHonorsDepth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Root</title></head><body><a href="/child">Child</a></body></html>`))
	})
	mux.HandleFunc("/child", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Child</title></head><body>Leaf</body></html>`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	one := newCrawler(true).crawl(server.URL, 1, nil)
	if one == nil {
		t.Fatal("depth-one crawl returned nil")
	}
	if len(one.Links) != 0 {
		t.Fatalf("depth-one crawl returned %d child links, want 0", len(one.Links))
	}

	two := newCrawler(true).crawl(server.URL, 2, nil)
	if two == nil {
		t.Fatal("depth-two crawl returned nil")
	}
	if len(two.Links) != 1 || two.Links[0] == nil || !strings.HasSuffix(two.Links[0].URL, "/child") {
		t.Fatalf("depth-two child links = %+v, want /child", two.Links)
	}
}

func TestCrawlHandlerRequiresTokenWhenConfigured(t *testing.T) {
	t.Setenv("VENI_API_TOKEN", "test-token")
	req := httptest.NewRequest(http.MethodGet, "/crawl?url=https%3A%2F%2Fexample.com", nil)
	rec := httptest.NewRecorder()
	crawlAuth(http.HandlerFunc(crawlHandler)).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated crawl status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/crawl?url=file%3A%2F%2F%2Fetc%2Fpasswd", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec = httptest.NewRecorder()
	crawlAuth(http.HandlerFunc(crawlHandler)).ServeHTTP(rec, req)
	if rec.Code == http.StatusUnauthorized {
		t.Fatal("authenticated crawl was rejected")
	}
}

func TestNonLoopbackListenRequiresToken(t *testing.T) {
	t.Setenv("VENI_API_TOKEN", "")
	if err := validateListenSecurity("0.0.0.0:8087"); err == nil {
		t.Fatal("non-loopback listener without token was accepted")
	}
	t.Setenv("VENI_API_TOKEN", "test-token")
	if err := validateListenSecurity("0.0.0.0:8087"); err != nil {
		t.Fatalf("non-loopback listener with token rejected: %v", err)
	}
}

func TestCrawlHandlerRejectsInvalidDepth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/crawl?url=https%3A%2F%2Fexample.com&depth=99", nil)
	w := httptest.NewRecorder()
	crawlHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
