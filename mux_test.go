package mux_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tombuente/mux"
)

func newTestMiddleware(key, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(key, value)
			next.ServeHTTP(w, r)
		})
	}
}

func respondEarlyMiddleware(_ http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMuxMiddleware(t *testing.T) {
	r := mux.New()
	r.Use(newTestMiddleware("X-Middleware-1", "1"))
	r.Use(newTestMiddleware("X-Middleware-2", "1"))

	r.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if header := rec.Header().Get("X-Middleware-1"); header != "1" {
		t.Errorf("Expected header X-Middleware to be \"1\", got \"%v\"", header)
	}
	if header := rec.Header().Get("X-Middleware-2"); header != "1" {
		t.Errorf("Expected header X-Middleware to be \"1\", got \"%v\"", header)
	}
}

func TestMuxMiddlewareExecOrder(t *testing.T) {
	r := mux.New()
	r.Use(respondEarlyMiddleware)
	r.Use(newTestMiddleware("X-Middleware", "1"))

	r.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if header := rec.Header().Get("X-Middleware"); header != "" {
		t.Errorf("Expected header X-Middleware to not be set because respondEarlyMiddleware should've returned early, got \"%v\"", header)
	}
}

func TestMuxNestedMiddleware(t *testing.T) {
	r := mux.New()
	r.Use(newTestMiddleware("X-Middleware", "1"))
	r.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	nested := mux.New()
	nested.Use(newTestMiddleware("X-Middleware-Nested", "1"))
	nested.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Mount("/nested", nested)

	rec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if header := rec.Header().Get("X-Middleware"); header != "1" {
		t.Errorf("Expected header X-Middleware to be \"1\", got \"%v\"", header)
	}

	req, err = http.NewRequest(http.MethodGet, "/nested/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
	if header := rec.Header().Get("X-Middleware"); header != "1" {
		t.Errorf("Expected header X-Middleware to be \"1\", got \"%v\"", header)
	}
	if header := rec.Header().Get("X-Middleware-Nested"); header != "1" {
		t.Errorf("Expected header X-Middleware-Nested to be \"1\", got \"%v\"", header)
	}
}

func TestMuxMount(t *testing.T) {
	r := mux.New()

	n1 := mux.New()
	n1.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	n2 := mux.New()
	n2.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Mount("/n1", n1)
	r.Mount("/n2/", n2)

	rec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/n1/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	req, err = http.NewRequest(http.MethodGet, "/n2/test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}
