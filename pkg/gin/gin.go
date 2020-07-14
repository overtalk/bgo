package ginpkg

import (
	"encoding/xml"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/xml"
)

type Config struct {
	XMLName  xml.Name `xml:"xml"`
	Name     string   `xml:"name"`
	Host     string   `xml:"host"`
	Port     int      `xml:"port"`
	CertFile string   `xml:"certFile"`
	KeyFile  string   `xml:"keyFile"`
}

type GinServer struct {
	cfg    *Config
	engine *gin.Engine
}

func NewGinServer(path string) (*GinServer, error) {
	gin.SetMode(gin.ReleaseMode)
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return nil, err
	}
	return &GinServer{engine: gin.New(), cfg: cfg}, nil
}

func (this *GinServer) GinEngine() *gin.Engine { return this.engine }

func (this *GinServer) Start() {
	go func() {
		var err error
		addr := fmt.Sprintf("%s:%d", this.cfg.Host, this.cfg.Port)
		logpkg.Info("start gin http server", zap.Any("name", this.cfg.Name))
		if len(this.cfg.CertFile) != 0 && len(this.cfg.CertFile) != 0 {
			err = this.engine.RunTLS(addr, this.cfg.CertFile, this.cfg.KeyFile)
		} else {
			err = this.engine.Run(addr)
		}
		if err != nil {
			logpkg.Fatal("start gin error", zap.Error(err), zap.Any("addr", addr))
		}
	}()
}
