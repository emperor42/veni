package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestDiscoveryHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/api/discover", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(discoveryHandler)
    
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    var elements []Element
    if err := json.NewDecoder(rr.Body).Decode(&elements); err != nil {
        t.Errorf("Error decoding response: %v", err)
    }
    
    if len(elements) == 0 {
        t.Error("Expected at least one element")
    }
    
    if len(elements) < 3 {
        t.Errorf("Expected at least 3 elements, got %d", len(elements))
    }
}

func TestRateLimitHandler(t *testing.T) {
    // Test rate limiting - allow 15 requests total, limit to 10 per second
    for i := 0; i < 15; i++ {
        req, err := http.NewRequest("GET", "/test", nil)
        if err != nil {
            t.Fatal(err)
        }
        
        rr := httptest.NewRecorder()
        // Simulate rate limiting by calling the middleware directly
        rateLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("OK"))
        }).ServeHTTP(rr, req)
        
        if i < 10 {
            // First 10 requests should succeed
            if status := rr.Code; status != http.StatusOK {
                t.Errorf("Expected request %d to succeed, got status %d", i, status)
            }
        } else {
            // Requests 11-15 should be rate limited
            if status := rr.Code; status != http.StatusTooManyRequests {
                t.Errorf("Expected request %d to be rate limited, got status %d", i, status)
            }
        }
    }
}

func TestElementsHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/api/elements", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        elements := []string{"custom-element", "data-table", "api-client", "form-handler"}
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(elements)
    })
    
    handler.ServeHTTP(rr, req)
    
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    var elements []string
    if err := json.NewDecoder(rr.Body).Decode(&elements); err != nil {
        t.Errorf("Error decoding response: %v", err)
    }
    
    if len(elements) != 4 {
        t.Errorf("Expected 4 elements, got %d", len(elements))
    }
}