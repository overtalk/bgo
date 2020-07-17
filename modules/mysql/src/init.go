package cmysql

import (
	"go.uber.org/zap"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/modules/mysql"
	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/pkg/mysql"
)

func init() {
	var module imysql.IMysqlModule = new(CMysqlModule)
	core.GetCore().RegisterModule(imysql.ModuleName, module)
}

type CMysqlModule struct {
	core.Module

	mysqlConn *mysqlpkg.MysqlConn
}

func (this *CMysqlModule) LoadConfig(path string) error {
	mysqlConn, err := mysqlpkg.NewMysqlConn(path)
	if err != nil {
		logpkg.Error("load config error", zap.Any("path", path), zap.Any("module", this.GetName()))
		return err
	}

	this.mysqlConn = mysqlConn
	return nil
}

func (this *CMysqlModule) PreTicker() error {
	return this.mysqlConn.Connect()
}
