package zaplog

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	fileutil "github.com/overtalk/bgo/utils/file"
)

func InitLogger(cfg *Config) {
	var encoder zapcore.Encoder

	if cfg.UseJsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var minLevel zapcore.Level
	if err := minLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		log.Fatalf("failed to parse log level : %s\n", cfg.LogLevel)
	}
	Level.SetLevel(minLevel)

	var cores []zapcore.Core
	if len(cfg.LogFilePath) > 0 {
		if !fileutil.PathExists(cfg.LogFilePath) {
			panic("log path " + cfg.LogFilePath + " is absent")
		}
		if !fileutil.IsDir(cfg.LogFilePath) {
			panic("log path " + cfg.LogFilePath + " should be a dir")
		}

		cfg.LogFilePath = filepath.Join(cfg.LogFilePath, cfg.LogFilePrefix+".log")

		hook, err := rotatelogs.New(
			strings.Replace(cfg.LogFilePath, ".log", "", -1)+"-%Y%m%d%H.log",
			rotatelogs.WithLinkName(cfg.LogFilePath),
			rotatelogs.WithMaxAge(time.Duration(cfg.MaxKeepHour)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(cfg.RotationHour)*time.Hour),
		)
		if err != nil {
			log.Fatalf("failed to build logger : %v\n", err)
		}

		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(hook), Level))
	}

	if cfg.OpenConsole || len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), Level))
	}

	zap.ReplaceGlobals(zap.New(zapcore.NewTee(cores...), zap.AddCaller()))
	return
}

var (
	Level         = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	encoderConfig = zapcore.EncoderConfig{
		TimeKey:      "ts",
		MessageKey:   "msg",
		LevelKey:     "lvl",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
)

func init() {
	InitLogger(&Config{
		LogFilePrefix: "default",
		LogLevel:      "debug",
		UseJsonFormat: false,
		LogFilePath:   "",
		OpenConsole:   true,
		MaxKeepHour:   24 * 7,
		RotationHour:  1,
	})
}

// Config logger configuration
type Config struct {
	LogFilePrefix string
	LogLevel      string
	UseJsonFormat bool   // weather to use json format
	LogFilePath   string // log file path
	OpenConsole   bool   // open the log in console
	MaxKeepHour   int
	RotationHour  int
}
