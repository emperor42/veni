package veni

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

// ResponseRecorder wraps an http.ResponseWriter to capture the response
// status code and body. This allows the middleware to inspect and modify
// the content before it is sent to the client.
//
// It buffers the entire response in memory, allowing the VENI middleware
// to parse the HTML, inject scripts, and then write the modified content.
type ResponseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

// NewResponseRecorder creates a new ResponseRecorder wrapping the provided writer.
// It initializes the status to 200 (OK) and prepares the buffer for body capture.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

// WriteHeader captures the status code and writes it to the underlying writer.
// It ensures WriteHeader is only called once, as per the HTTP spec.
func (r *ResponseRecorder) WriteHeader(status int) {
	if r.status != http.StatusOK {
		return // Header already written
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write captures the response body into the internal buffer.
// It does NOT write to the underlying ResponseWriter immediately.
// The data is held in memory until Flush() is called.
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

// Status returns the captured HTTP status code.
func (r *ResponseRecorder) Status() int {
	return r.status
}

// Body returns a copy of the captured response body.
func (r *ResponseRecorder) Body() []byte {
	return r.body.Bytes()
}

// Flush writes the captured body to the underlying ResponseWriter.
// This is the primary method to send the response to the client.
// It ensures headers are written if they haven't been yet.
func (r *ResponseRecorder) Flush() {
	// Ensure headers are written if they haven't been
	if r.status != http.StatusOK {
		r.ResponseWriter.WriteHeader(r.status)
	}

	// Write the captured body
	if r.body.Len() > 0 {
		r.ResponseWriter.Write(r.body.Bytes())
	}
}

// Hijack implements the http.Hijacker interface to support WebSocket upgrades
// and other protocols that require direct connection access.
// If the underlying ResponseWriter supports Hijacking, we delegate to it.
func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Ensure ResponseRecorder implements the necessary interfaces
var (
	_ http.ResponseWriter = (*ResponseRecorder)(nil)
	_ http.Hijacker       = (*ResponseRecorder)(nil)
	_ http.Flusher        = (*ResponseRecorder)(nil)
)
