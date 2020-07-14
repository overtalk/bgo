package cpprof

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/internal/pprof"
	"github.com/overtalk/bgo/pkg/log"
)

func init() {
	var module ipprof.IPProf = new(CPProf)
	core.GetCore().RegisterModule(ipprof.ModuleName, module)
}

type CPProf struct {
	core.Module

	ip   string
	port int
	mux  *http.ServeMux
	svr  *http.Server
}

func (this *CPProf) Init() error {
	this.ip = "127.0.0.1"
	this.port = 9000
	this.mux = http.NewServeMux()
	this.svr = &http.Server{
		Handler: this.mux,
	}
	return nil
}

func (this *CPProf) PreTicker() error {
	for partten, handler := range map[string]http.HandlerFunc{
		"/debug/pprof/":        pprof.Index,
		"/debug/pprof/cmdline": pprof.Cmdline,
		"/debug/pprof/profile": pprof.Profile,
		"/debug/pprof/symbol":  pprof.Symbol,
		"/debug/pprof/trace":   pprof.Trace,
	} {
		this.mux.Handle(partten, handler)
	}

	this.svr.Addr = fmt.Sprintf("%s:%d", this.ip, this.port)

	go func() {
		logpkg.Info("start pprof http server", zap.Any("addr", this.svr.Addr))
		if err := this.svr.ListenAndServe(); err != nil {
			logpkg.Fatal("start pprof http server error", zap.Error(err))
		}
	}()

	return nil
}

func (this *CPProf) SetPProfPort(ip string, port int) {
	this.ip = ip
	this.port = port
}
