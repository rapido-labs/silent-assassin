package logger

import (
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const EVENT = "event"
const COMPONENT = "component"

type IZapLogger interface {
	Debug(message string, event string, component string)
	Info(message string, event string, component string)
	Error(message string, event string, component string)
	Warn(message string, event string, component string)
}

type ZapLogger struct {
	*zap.Logger
}

func (l ZapLogger) Debug(message string, event string, component string) {
	l.Logger.Debug(message, zap.String(EVENT, event), zap.String(COMPONENT, component))
}

func (l ZapLogger) Info(message string, event string, component string) {
	l.Logger.Info(message, zap.String(EVENT, event), zap.String(COMPONENT, component))
}

func (l ZapLogger) Warn(message string, event string, component string) {
	l.Logger.Warn(message, zap.String(EVENT, event), zap.String(COMPONENT, component))
}

func (l ZapLogger) Error(message string, event string, component string) {
	l.Logger.Error(message, zap.String(EVENT, event), zap.String(COMPONENT, component))
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
