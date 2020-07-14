package caliyunoss

import (
	"encoding/xml"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/internal/aliyun/oss"
	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/xml"
)

type Config struct {
	XMLName          xml.Name `xml:"xml"`
	Ak               string   `xml:"ak"`
	Sk               string   `xml:"sk"`
	Endpoint         string   `xml:"endpoint"`
	InternetEndpoint string   `xml:"internetEndpoint"`
	Bucket           string   `xml:"bucket"`
	Dir              string   `xml:"dir"`
}

func (this *CAliyunOssModule) LoadConfig(path string) error {
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return err
	}
	this.cfg = cfg

	logpkg.Info("config", zap.Any("module", ialiyunoss.ModuleName), zap.Any("config", cfg))
	return nil
}
