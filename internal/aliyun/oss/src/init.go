package caliyunoss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/internal/aliyun/oss"
)

func init() {
	var module ialiyunoss.IAliyunOssModule = new(CAliyunOssModule)
	core.GetCore().RegisterModule(ialiyunoss.ModuleName, module)
}

type CAliyunOssModule struct {
	core.Module

	cfg *Config

	ossClient *oss.Client
	ossBucket *oss.Bucket
}

func (this *CAliyunOssModule) Init() error {
	cli, err := oss.New(this.cfg.InternetEndpoint, this.cfg.Ak, this.cfg.Sk)
	if err != nil {
		return err
	}

	b, err := cli.Bucket(this.cfg.Bucket)
	if err != nil {
		return err
	}

	this.ossClient = cli
	this.ossBucket = b
	return nil
}
