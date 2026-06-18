// VENI – a tiny HTTP gateway that forwards GET requests to a static file
// server and turns specially‑crafted POST requests into arbitrary outbound
// HTTP calls.
//
//   • Middleware can be added by wrapping the main handler with any
//     http.Handler (the `useMiddleware` helper shows a simple example).
//   • Only GET and POST are accepted – everything else receives 405.
//   • The directory served by the built‑in file server is taken from the
//     environment variable `VENI`.
//   • POST bodies must contain:
//        - VENI_METHOD   – the HTTP method to use for the outbound request
//        - VENI_URI      – the target URI for the outbound request
//        - zero‑or‑more keys of the form VENI_HEADER__<HeaderName>
//        - optional key VENI_BODY containing the request payload
//
//   The server builds the outbound request, sends it, and streams the
//   response back to the original client.
//
// Build & run:
//
//   $ export VENI=/path/to/static/files
//   $ go build -o veni main.go
//   $ ./veni
//
package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ---------------------------------------------------------------------
// Helper: simple middleware chaining
// ---------------------------------------------------------------------

// middleware is a function that takes and returns an http.Handler.
type middleware func(http.Handler) http.Handler

// useMiddleware applies a slice of middlewares to a final handler.
func useMiddleware(final http.Handler, mws ...middleware) http.Handler {
	h := final
	// Apply in reverse order so the first middleware in the slice runs first.
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Example middleware that logs each request.
func loggingMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------
// Core handler
// ---------------------------------------------------------------------

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// -----------------------------------------------------------------
	// 1️⃣ Enforce allowed methods (GET, POST)
	// -----------------------------------------------------------------
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// -----------------------------------------------------------------
	// 2️⃣ GET → serve static files from $VENI
	// -----------------------------------------------------------------
	if r.Method == http.MethodGet {
		dir := os.Getenv("VENI")
		if dir == "" {
			http.Error(w, "Server mis‑configuration: VENI env var not set", http.StatusInternalServerError)
			return
		}
		fs := http.FileServer(http.Dir(dir))
		// Strip any leading slash to keep the file server rooted at dir.
		http.StripPrefix("/", fs).ServeHTTP(w, r)
		return
	}

	// -----------------------------------------------------------------
	// 3️⃣ POST → build and forward a new request
	// -----------------------------------------------------------------
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MiB max memory
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ----- Required fields -------------------------------------------------
	venMethod := r.FormValue("VENI_METHOD")
	if venMethod == "" {
		http.Error(w, "Missing required field VENI_METHOD", http.StatusBadRequest)
		return
	}
	venURI := r.FormValue("VENI_URI")
	if venURI == "" {
		http.Error(w, "Missing required field VENI_URI", http.StatusBadRequest)
		return
	}

	// Validate that the supplied method is a known HTTP verb.
	allowedMethods := map[string]bool{
		http.MethodGet:    true,
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
		http.MethodDelete: true,
		http.MethodHead:   true,
		http.MethodOptions:true,
	}
	if !allowedMethods[strings.ToUpper(venMethod)] {
		http.Error(w, "Unsupported VENI_METHOD: "+venMethod, http.StatusBadRequest)
		return
	}

	// ----- Optional body ----------------------------------------------------
	venBody := r.FormValue("VENI_BODY")
	var bodyReader io.Reader
	if venBody != "" {
		bodyReader = strings.NewReader(venBody)
	}

	// ----- Headers ---------------------------------------------------------
	venHeaders := make(map[string]string)
	for key, values := range r.PostForm {
		// Look for keys that start with "VENI_HEADER__"
		if strings.HasPrefix(key, "VENI_HEADER__") {
			headerName := strings.TrimPrefix(key, "VENI_HEADER__")
			// In case the same header appears multiple times we join them with commas.
			venHeaders[headerName] = strings.Join(values, ",")
		}
	}

	// -----------------------------------------------------------------
	// 4️⃣ Construct outbound request
	// -----------------------------------------------------------------
	outReq, err := http.NewRequest(strings.ToUpper(venMethod), venURI, bodyReader)
	if err != nil {
		http.Error(w, "Failed to create outbound request: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Populate headers collected from the form.
	for k, v := range venHeaders {
		outReq.Header.Set(k, v)
	}
	// If a body is present and the caller didn't set a Content-Type, default to text/plain.
	if venBody != "" && outReq.Header.Get("Content-Type") == "" {
		outReq.Header.Set("Content-Type", "text/plain")
	}

	// -----------------------------------------------------------------
	// 5️⃣ Execute request (using default transport)
	// -----------------------------------------------------------------
	client := http.DefaultClient
	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, "Outbound request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// -----------------------------------------------------------------
	// 6️⃣ Relay response back to the original caller
	// -----------------------------------------------------------------
	// Copy status code.
	w.WriteHeader(resp.StatusCode)
	// Copy all response headers (except hop‑by‑hop ones that net/http strips automatically).
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	// Stream body.
	if _, err := io.Copy(w, resp.Body); err != nil {
		// We can't really do much here; log and finish.
		log.Printf("error streaming response body: %v", err)
	}
}

// ---------------------------------------------------------------------
// Server bootstrap
// ---------------------------------------------------------------------

func main() {
	// Build the final handler with any middleware you want.
	// Add or remove middleware functions in the slice below.
	handler := useMiddleware(http.HandlerFunc(mainHandler),
		loggingMw, // <-- example logging middleware
		// You could add more, e.g., authentication, rate‑limiting, etc.
	)

	// Listen on port 8080 (feel free to change).
	const addr = ":8080"
	log.Printf("VENI listening on %s – static root from $VENI", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
