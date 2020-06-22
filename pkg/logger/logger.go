package logger

import (
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const EVENT = "event"
const COMPONENT = "component"

type IZapLogger interface {
	Debug(message string)
	Info(message string)
	Error(message string)
	Warn(message string)
}

type ZapLogger struct {
	*zap.Logger
}

func (l ZapLogger) Debug(message string) {
	l.Logger.Debug(message)
}

func (l ZapLogger) Info(message string) {
	l.Logger.Info(message)
}

func (l ZapLogger) Warn(message string) {
	l.Logger.Warn(message)
}

func (l ZapLogger) Error(message string) {
	l.Logger.Error(message)
}

func Init(cp config.IProvider) ZapLogger {

	config := zap.Config{
		Encoding:         "json",
		Level:            getLogLevel(cp.GetString(config.LogLevel)),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "timestamp",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
	}

	logger, _ := config.Build()
	defer logger.Sync()

	return ZapLogger{Logger: logger}
}

func getLogLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}
