package mux

import "net/http"

var _ http.Handler = (*Mux)(nil)

type Mux struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

// New returns a new Mux.
func New() *Mux {
	return &Mux{
		mux:         http.NewServeMux(),
		middlewares: []func(http.Handler) http.Handler{},
	}
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Use appends middleswares to the middlware stack.
func (m *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	m.middlewares = append(m.middlewares, middlewares...)
}

// Handle registers the handler for the given pattern.
// If the given pattern conflicts, with one that is already registered,
// Handle panics.
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.register(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
// If the given pattern conflicts, with one that is already registered,
// HandleFunc panics.
func (m *Mux) HandleFunc(pattern string, handleFunc http.HandlerFunc) {
	m.register(pattern, http.HandlerFunc(handleFunc))
}

// Mount attaches an http.Handler along pattern.
// It appends a tailing slash if none is present.
// The http.Handler is mounted with http.StripPrefix.
func (m *Mux) Mount(pattern string, handler http.Handler) {
	// Trim tailing slash
	if pattern[len(pattern)-1] == '/' {
		pattern = pattern[:len(pattern)-1]
	}

	m.Handle(pattern+"/", http.StripPrefix(pattern, handler))
}

// Route mounts a mux on pattern
func (m *Mux) Route(pattern string, fn func(mux *Mux)) {
	mux := New()
	fn(mux)
	m.Mount(pattern, mux)
}

func (m *Mux) register(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, m.wrapMiddleware(handler))
}

func (m *Mux) wrapMiddleware(handler http.Handler) http.Handler {
	for i := range m.middlewares {
		handler = m.middlewares[len(m.middlewares)-1-i](handler)
	}

	return handler
}
