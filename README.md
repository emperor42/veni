VENI – Go HTTP Gateway & Static File Server

A lightweight Go server that:

    Serves static files from a directory defined by the VENI environment variable.
    Accepts only GET and POST requests (all others receive 405 Method Not Allowed).
    Provides a simple middleware‑stack mechanism for easy extensibility.
    Turns specially‑crafted POST requests into arbitrary outbound HTTP calls, allowing you to specify method, URL, headers, and body via form data.

Table of Contents

    Features
    Prerequisites
    Installation & Build
    Configuration
    Running the Server
    API Details
        GET – Static File Serving
        POST – Proxy / Forwarding
    Middleware Architecture
    Extending the Server
    Logging & Observability
    Error Handling
    Security Considerations
    License

Features

    Zero‑dependency – uses only the Go standard library.
    Configurable static root via the VENI environment variable.
    Method whitelisting – only GET and POST are accepted.
    Dynamic outbound request generation from POST form data (VENI_METHOD, VENI_URI, VENI_HEADER__*, VENI_BODY).
    Middleware-friendly – plug‑in logging, auth, rate‑limiting, etc., without touching core logic.
    Streaming responses – the server pipes the remote response straight back to the original client.

Prerequisites

    Go 1.22 or newer installed (any recent version works).
    A directory containing the static assets you wish to serve.

Installation & Build

# Clone or copy the source into a folder, e.g., veni/
cd veni

# Build the binary
go build -o veni main.go

The resulting executable (veni) is ready to run.
Configuration
Variable	Description	Example
VENI	Absolute path to the directory that the built‑in file server will expose.	export VENI=/var/www/public

    Important: The server will refuse to start if VENI is unset.

Running the Server

# Set the static directory
export VENI=/path/to/static/files

# Start the server (listens on :8080 by default)
./veni

You should see a log line similar to:

2025/11/11 14:32:01 VENI listening on :8080 – static root from $VENI

Visit http://localhost:8080 in a browser to browse the static files.
API Details
GET – Static File Serving

    Endpoint: /* (any path)
    Behaviour: The request is handed to http.FileServer rooted at $VENI.
    Restrictions: Only GET is permitted; any other method receives 405.

POST – Proxy / Forwarding

A POST request can trigger an outbound HTTP call based on its form data.
Form Field	Required?	Description
VENI_METHOD	✅	HTTP method for the outbound request (e.g., GET, POST, PUT, DELETE, …).
VENI_URI	✅	Full target URI (including scheme, host, path, query).
VENI_HEADER__<HeaderName>	❌	One or more custom headers. Replace <HeaderName> with the exact header name you want to send. Example: VENI_HEADER__Authorization: Bearer xyz.
VENI_BODY	❌	Raw request body for the outbound call (sent as plain text unless a Content-Type header is supplied via VENI_HEADER__Content-Type).
Request Flow

    Validate presence of VENI_METHOD and VENI_URI.
    Collect any VENI_HEADER__* entries into a map.
    Create a new http.Request using the supplied method, URI, headers, and optional body.
    Execute the request with the default HTTP client.
    Relay the remote response status, headers, and body back to the original caller.

Example cURL

curl -X POST http://localhost:8080 \
  -F "VENI_METHOD=POST" \
  -F "VENI_URI=https://httpbin.org/post" \
  -F "VENI_HEADER__Content-Type=application/json" \
  -F "VENI_HEADER__X-Custom-Header=demo" \
  -F 'VENI_BODY={"msg":"hello"}'

The server will forward the request to https://httpbin.org/post and stream the response back.
Middleware Architecture

The server ships with a tiny middleware helper:

type middleware func(http.Handler) http.Handler
func useMiddleware(final http.Handler, mws ...middleware) http.Handler

You can wrap the core handler with any number of middleware functions.
An example logging middleware (loggingMw) is already included.
Adding Your Own Middleware

func myAuthMw(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Perform auth checks…
        next.ServeHTTP(w, r)
    })
}

// Register:
handler := useMiddleware(http.HandlerFunc(mainHandler),
    loggingMw,
    myAuthMw,
)

Extending the Server

    Change listening port – edit the addr constant in main().
    TLS support – replace http.ListenAndServe with http.ListenAndServeTLS and provide certificate files.
    Rate limiting, CORS, tracing, etc. – implement as middleware and insert into the stack.
    Custom error pages – modify the http.Error calls or render HTML templates.

Logging & Observability

The built‑in loggingMw prints:

<client IP> <METHOD> <PATH>

You can replace it with structured logging (e.g., logrus, zap) if desired. All other errors are logged via the standard log package.
Error Handling

    Missing VENI env var → 500 Internal Server Error with a clear message.
    Invalid/missing form fields → 400 Bad Request.
    Unsupported outbound method → 400 Bad Request.
    Failure contacting the target URI → 502 Bad Gateway.

All error messages are deliberately concise to avoid leaking internal details.
Security Considerations

    Input validation – only known HTTP verbs are allowed for VENI_METHOD.
    Header injection – header names are taken verbatim from the form key after VENI_HEADER__. Ensure callers are trusted or add additional validation if exposing publicly.
    Potential open‑proxy misuse – if the server is reachable from the internet, anyone can cause it to issue arbitrary outbound requests. Deploy behind authentication or restrict network access as needed.
    Static file exposure – the server serves exactly what resides under $VENI. Do not point it at sensitive directories.

License

MIT License – feel free to modify, redistribute, and use in commercial projects.

Enjoy building with VENI! If you encounter any issues or have ideas for improvements, feel free to open a pull request or submit an issue on the project's repository.
