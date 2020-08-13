package zaplog_test

import (
	"testing"

	"go.uber.org/zap"

	_ "github.com/overtalk/bgo/pkg/zaplog"
)

func TestInitLogger(t *testing.T) {
	zap.L().Info("xsdf")
}
