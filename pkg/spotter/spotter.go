package spotter

import (
	"fmt"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
)

func Start(cp *config.Provider, zl logger.ZapLogger) {

	zl.Debug(logger.Log{Event: "START", Component: "SILENT_ASSASSIN", Message: fmt.Sprintf("Starting Spotter Loop with a delay interval of %s", cp.GetString("spotter.poll_interval_ms"))})
	// zl.Debug(logger.Log{Event: "START", Component: "SILENT_ASSASSIN", Message: ""})
}
