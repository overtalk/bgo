package route

import (
	"time"
)

// IRouteEnabler enable or disable some routes
type IRouteEnabler interface {
	Enabled(uint8, uint8) bool
}

type fullRouteEnabler struct{}

func (*fullRouteEnabler) Enabled(_ uint8, _ uint8) bool { return true }

// FullRouteEnabler a fullRouteEnabler struct
var FullRouteEnabler IRouteEnabler = &fullRouteEnabler{}

// ITimeouter wait a while and return a timeout proto.Message
type ITimeouter interface {
	Timeout() time.Duration
	Result() IOutProtocol
}

// Router a module router
type Router struct {
	modules  map[uint8]IModule
	enabler  IRouteEnabler
	timeout  ITimeouter
	noneResp IOutProtocol
}

// RouterOptionFunc set the Router's option
type RouterOptionFunc func(*Router)

// OptionRouteEnabler set Router's enabler
func OptionRouteEnabler(enabler IRouteEnabler) RouterOptionFunc {
	return func(r *Router) {
		r.enabler = enabler
	}
}

// OptionTimeoutResponse set Router's timeout
func OptionTimeoutResponse(timeout ITimeouter) RouterOptionFunc {
	return func(r *Router) {
		r.timeout = timeout
	}
}

// OptionNoneResponse set Router's noneResp
func OptionNoneResponse(resp IOutProtocol) RouterOptionFunc {
	return func(r *Router) {
		r.noneResp = resp
	}
}

// NewRouter create a Router struct
func NewRouter(opts ...RouterOptionFunc) *Router {
	router := &Router{map[uint8]IModule{}, FullRouteEnabler, nil, nil}
	for _, opt := range opts {
		opt(router)
	}
	return router
}

// Register register several modules
func (router *Router) Register(modules ...IModule) {
	for _, m := range modules {
		router.modules[m.GetMID()] = m
	}
}

// Dispatch dispatch each client's request
func (router *Router) Dispatch(r IRequest) (IOutProtocol, bool) {
	moduleID := r.GetMID()
	actionID := r.GetAID()

	var module IModule
	if router.enabler.Enabled(moduleID, actionID) {
		var ok bool
		module, ok = router.modules[moduleID]
		if !ok {
			//TODO: log
			//zaplog.S.Errorf("router: module(%d) not found", moduleID)
			return router.noneResp, false
		}
	} else {
		//TODO: log
		//zaplog.S.Errorf(
		//	"router: module(%d) action(%d) disabled", moduleID, actionID)
		return router.noneResp, false
	}
	if router.timeout == nil {
		return module.Handle(r), false
	}
	// timeout to handle a request
	result := make(chan IOutProtocol, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//TODO: log
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()
		result <- module.Handle(r)
	}()
	select {
	case pb := <-result:
		return pb, false
	case <-time.After(router.timeout.Timeout()):
		return router.timeout.Result(), true
	}
}
