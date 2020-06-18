package logger

import (
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const EVENT = "event"
const COMPONENT = "component"

type Log struct {
	Event     string
	Message   string
	Component string
}

type IZapLogger interface {
	Debug(log Log)
	Info(log Log)
	Error(log Log)
	Warn(log Log)
}

type ZapLogger struct {
	logger *zap.Logger
}

func (l ZapLogger) Debug(log Log) {
	l.logger.Debug(log.Message, zap.String(EVENT, log.Event), zap.String(COMPONENT, log.Component))
}

func (l ZapLogger) Info(log Log) {
	l.logger.Info(log.Message, zap.String(EVENT, log.Event), zap.String(COMPONENT, log.Component))
}

func (l ZapLogger) Warn(log Log) {
	l.logger.Warn(log.Message, zap.String(EVENT, log.Event), zap.String(COMPONENT, log.Component))
}

func (l ZapLogger) Error(log Log) {
	l.logger.Error(log.Message, zap.String(EVENT, log.Event), zap.String(COMPONENT, log.Component))
}

func Init(cp *config.Provider) ZapLogger {

	config := zap.Config{
		Encoding:         "json",
		Level:            getLogLevel(cp.GetString("LOGGER.LEVEL")),
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

	return ZapLogger{logger: logger}
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
