package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Element struct {
    Tag       string   `json:"tag"`
    URL       string   `json:"url"`
    Version   string   `json:"version"`
    Functions []string `json:"functions"`
}

// Rate limiter - simple token bucket
var requests = make(chan string, 100)
var limiter = make(chan bool, 10)

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        select {
        case requests <- r.RemoteAddr:
            next(w, r)
        default:
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
        }
    }
}

func discoveryHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    elements := []Element{
        {
            Tag: "custom-element",
            URL: "/elements/custom-element.js",
            Functions: []string{"render", "behave"},
        },
        {
            Tag: "data-table",
            URL: "/elements/data-table.js",
            Functions: []string{"fetchData", "renderTable"},
        },
        {
            Tag: "api-client",
            URL: "/elements/api-client.js",
            Functions: []string{"callAPI", "handleResponse"},
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(elements); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
    // Security middleware
    handler := http.NewServeMux()
    
    // Health endpoint
    handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        response := map[string]interface{}{
            "status": "healthy",
            "service": "veni",
            "version": "1.0.0",
            "timestamp": "2024-01-01T00:00:00Z",
        }
        json.NewEncoder(w).Encode(response)
    })
    
    // Discovery endpoint with rate limiting
    handler.HandleFunc("/api/discover", rateLimitMiddleware(discoveryHandler))
    
    // Additional endpoints
    handler.HandleFunc("/api/elements", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        elements := []string{"custom-element", "data-table", "api-client", "form-handler"}
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(elements)
    })
    
    // Start server
    fmt.Printf("VENI server starting on port 8082\n")
    if err := http.ListenAndServe(":8082", handler); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
}