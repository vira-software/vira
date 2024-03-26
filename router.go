package vira

import (
	"fmt"
	"net/http"
	"strings"
)

// Params represents a map of URL parameters
type Params map[string]string

// HandlerFunc is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type HandlerFunc func(http.ResponseWriter, *http.Request, Params)

// Router is a tire base HTTP request router which can be used to
// dispatch requests to different handler functions.
type Router struct {
	trie      *Trie
	otherwise HandlerFunc
}

// NewRouter retuns a Router instance.
func NewRouter(ops ...Options) *Router {
	return &Router{trie: New(ops...)}
}

// Get registers a new Get route for a path with matching handler in the Router.
func (r *Router) Get(endpoint string, handler HandlerFunc) {
	r.Handle(http.MethodGet, endpoint, handler)
}

// Head registers a new HEAD route for a path with matching handler in the Router.
func (r *Router) Head(endpoint string, handler HandlerFunc) {
	r.Handle(http.MethodHead, endpoint, handler)
}

// Post registers a new POST route for a path with matching handler in the Router.
func (m *Router) Post(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPost, pattern, handler)
}

// Put registers a new PUT route for a path with matching handler in the Router.
func (m *Router) Put(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPut, pattern, handler)
}

// Patch registers a new PATCH route for a path with matching handler in the Router.
func (m *Router) Patch(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodPatch, pattern, handler)
}

// Delete registers a new DELETE route for a path with matching handler in the Router.
func (m *Router) Delete(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodDelete, pattern, handler)
}

// Options registers a new OPTIONS route for a path with matching handler in the Router.
func (m *Router) Options(pattern string, handler HandlerFunc) {
	m.Handle(http.MethodOptions, pattern, handler)
}

// Otherwise registers a new handler in the Router
// that will run if there is no other handler matching.
func (m *Router) Otherwise(handler HandlerFunc) {
	m.otherwise = handler
}

// Handle registers a new handler with method and path in the Router.
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, pattern string, handler HandlerFunc) {
	if method == "" {
		panic(fmt.Errorf("invalid method"))
	}
	r.trie.Define(pattern).Handle(strings.ToUpper(method), handler)
}

// Handler is an adapter which allows the usage of an http.Handler as a
// request handle.
func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path, func(w http.ResponseWriter, req *http.Request, _ Params) {
		handler.ServeHTTP(w, req)
	})
}

// HandlerFunc is an adapter which allows the usage of an http.HandlerFunc as a
// request handle.
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handler(method, path, handler)
}

// ServeHTTP implemented http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler HandlerFunc
	path := req.URL.Path
	method := req.Method
	res := r.trie.Match(path)

	if res.Node == nil {
		// FixedPathRedirect or TrailingSlashRedirect
		if res.TSR != "" || res.FPR != "" {
			req.URL.Path = res.TSR
			if res.FPR != "" {
				req.URL.Path = res.FPR
			}
			code := 301
			if method != "GET" {
				code = 307
			}
			http.Redirect(w, req, req.URL.String(), code)
			return
		}

		if r.otherwise == nil {
			http.Error(w, fmt.Sprintf(`"%s" not implemented`, path), 501)
			return
		}
		handler = r.otherwise
	} else {
		ok := false
		if handler, ok = res.Node.GetHandler(method).(HandlerFunc); !ok {
			// OPTIONS support
			if method == http.MethodOptions {
				w.Header().Set("Allow", res.Node.GetAllow())
				w.WriteHeader(204)
				return
			}

			if r.otherwise == nil {
				// If no route handler is returned, it's a 405 error
				w.Header().Set("Allow", res.Node.GetAllow())
				http.Error(w, fmt.Sprintf(`"%s" not allowed in "%s"`, method, path), 405)
				return
			}
			handler = r.otherwise
		}
	}

	handler(w, req, res.Params)
}
