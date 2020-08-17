package route

import "net/http"

// -----------------------------------------------
// httpModule
// -----------------------------------------------

// IHTTPModule support http route
type IHTTPModule interface {
	http.Handler
	GetPath() string
}

// IHTTPRouteEnabler a http route enabler
type IHTTPRouteEnabler interface {
	Enabled(uri string) bool
}

type fullHTTPRouteEnabler struct{}

func (*fullHTTPRouteEnabler) Enabled(_ string) bool { return true }

// FullHTTPRouteEnabler a fullHTTPRouteEnabler
var FullHTTPRouteEnabler IHTTPRouteEnabler = &fullHTTPRouteEnabler{}

// HTTPRouter a http router
type HTTPRouter struct {
	*http.ServeMux
}

var _ http.Handler = (*HTTPRouter)(nil)

// NewHTTPRouter create a HTTPRouter struct
func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{http.NewServeMux()}
}

// Register register several modules
func (r *HTTPRouter) Register(modules ...IHTTPModule) {
	for _, m := range modules {
		r.Handle(m.GetPath(), m)
	}
}
