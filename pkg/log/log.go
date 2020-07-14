package logpkg

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/overtalk/bgo/utils/file"
)

func init() {
	flag.IntVar(&rotationHour, "r", 1, "log rotation hour")
	flag.IntVar(&maxKeepHour, "max", 24*7, "log max keep hour")
	flag.BoolVar(&useJsonFormat, "json", false, "log format json")
	flag.BoolVar(&openConsole, "console", true, "show log in console")
	flag.StringVar(&logFilePath, "logpath", "", "log save dir")
	flag.StringVar(&logLevel, "loglevel", "debug", "log level {debug|info|error|fatal}")
	flag.StringVar(&logFilePrefix, "prefix", "default", "log prefix")
}

var (
	once   sync.Once
	logger *zap.Logger

	logFilePrefix string
	logLevel      string
	useJsonFormat bool   // weather to use json format
	logFilePath   string // log file path
	openConsole   bool   // open the log in console
	maxKeepHour   int
	rotationHour  int

	encoderConfig = zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	}
)

func BuildLogger() {
	var encoder zapcore.Encoder

	if useJsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var minLevel zapcore.Level
	if err := minLevel.UnmarshalText([]byte(logLevel)); err != nil {
		log.Fatalf("failed to parse log level : %s\n", logLevel)
	}

	level := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= minLevel
	})

	var cores []zapcore.Core
	if len(logFilePath) > 0 {
		if !fileutil.PathExists(logFilePath) {
			panic("log path " + logFilePath + " is absent")
		}
		if !fileutil.IsDir(logFilePath) {
			panic("log path " + logFilePath + " should be a dir")
		}

		logFilePath = filepath.Join(logFilePath, logFilePrefix+".log")

		hook, err := rotatelogs.New(
			strings.Replace(logFilePath, ".log", "", -1)+"-%Y%m%d%H.log",
			rotatelogs.WithLinkName(logFilePath),
			rotatelogs.WithMaxAge(time.Duration(maxKeepHour)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(rotationHour)*time.Hour),
		)
		if err != nil {
			log.Fatalf("failed to build logger : %v\n", err)
		}

		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(hook), level))
	}

	if openConsole || len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
}

//
//func Log(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
//	switch lvl {
//	case zapcore.DebugLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Debug(msg)
//	case zapcore.InfoLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Info(msg)
//	case zapcore.WarnLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Warn(msg)
//	case zapcore.ErrorLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Error(msg)
//	case zapcore.DPanicLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).DPanic(msg)
//	case zapcore.PanicLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Panic(msg)
//	case zapcore.FatalLevel:
//		GetLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Fatal(msg)
//	}
//}

func getLogger() *zap.Logger {
	once.Do(func() {
		if logger == nil {
			BuildLogger()
		}
	})
	return logger
}

func Debug(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Debug(msg)
}
func Info(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Info(msg)
}
func Warn(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Warn(msg)
}
func Error(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Error(msg)
}
func DPanic(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).DPanic(msg)
}
func Panic(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Panic(msg)
}
func Fatal(msg string, fields ...zapcore.Field) {
	getLogger().WithOptions(zap.AddCallerSkip(1)).With(fields...).Fatal(msg)
}
