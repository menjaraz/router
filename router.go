package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Middleware ...
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Routing ...
type Routing struct {
	Router     *http.ServeMux
	prefixes   []string
	middleware []Middleware
	isGroup    bool
	paths      []string
}

// New ...
func New(r *http.ServeMux) *Routing {
	return &Routing{Router: r}
}

// Get ...
func (r *Routing) Get(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodGet)
}

// Head ...
func (r *Routing) Head(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodHead)
}

// Post ...
func (r *Routing) Post(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodPost)
}

// Put ...
func (r *Routing) Put(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodPut)
}

// Patch ...
func (r *Routing) Patch(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodPatch)
}

// Delete ...
func (r *Routing) Delete(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodDelete)
}

// Option ...
func (r *Routing) Option(uri string, action http.HandlerFunc, name ...string) {
	r.compose(uri, action, http.MethodOptions)
}

// HandleFilesystem ...
func (r *Routing) HandleFilesystem(uri string, handler http.Handler) {
	r.Router.Handle(uri, handler)
}

// Middleware provide a convenient mechanism for filtering HTTP requests entering your application.
func (r *Routing) Middleware(middleware ...Middleware) *Routing {
	r.middleware = middleware
	return r
}

func (r *Routing) compose(uri string, action http.HandlerFunc, method string) {

	//Grouping middleware
	wrapped := action
	for i := len(r.middleware) - 1; i >= 0; i-- {
		wrapped = r.middleware[i](wrapped)
	}

	url := uri
	if len(url) > 0 && url != "/" {
		url = fmt.Sprintf("/%s", uri)
	}

	if url == "/" {
		url = ""
	}

	pattern := fmt.Sprintf("%s %s", method, url)
	if len(r.prefixes) > 0 {
		// collect the prefixes
		var mergePrefix = fmt.Sprintf("/%s", strings.Join(r.prefixes, "/"))
		pattern = fmt.Sprintf("%s %s%s", method, mergePrefix, url)
		r.Router.HandleFunc(pattern, wrapped)
	} else {
		r.Router.HandleFunc(pattern, wrapped)
	}

	r.registerPattern(pattern)

	//After the route has been configured, if we're outside of group Router then we'll clean up the middleware
	//then when the user try to add a new Router and wan use a middleware then they have to specify again.
	if len(r.prefixes) == 0 && !r.isGroup {
		r.middleware = nil
	}
}

// Prefix method may be used to prefix each route in the group with a given URI.
// For example, you may want to prefix all route URIs within the group with admin:
func (r *Routing) Prefix(prefix string, f func()) {

	defer func() {
		if len(r.prefixes) > 0 {
			r.prefixes = r.prefixes[:len(r.prefixes)-1]
		}
	}()

	if len(prefix) == 0 {
		panic("Prefix(): the prefix can't be empty")
	}

	r.prefixes = append(r.prefixes, prefix)

	f()
}

// Group allow you to share route attributes, such as middleware or namespaces
// across a large number of routes without needing to define those
// attributes on each individual route. Shared attributes are
// specified in an slice format as the first parameter
// to the Route::group method.
func (r *Routing) Group(f func()) {
	defer func() {
		r.isGroup = false
	}()

	r.isGroup = true

	f()
}

func (r *Routing) registerPattern(pattern string) {
	r.paths = append(r.paths, pattern)
}

func (r *Routing) Paths() []string {
	return r.paths
}
