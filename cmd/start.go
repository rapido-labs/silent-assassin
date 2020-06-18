package cmd

import (
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/spotter"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the killer",
	Long: `Starts the killer. It has 2 components. One to assign the expiry time to preemtible nodes and one to
	gracefully kill the nodes which have outlived the expiry time.`,
	Run: func(cmd *cobra.Command, args []string) {
		configProvider := config.Init(cfgFile)
		zapLogger = logger.Init(configProvider)

		spotter.Start(configProvider, zapLogger)
	},
}

func init() {
	//
	rootCmd.AddCommand(startCmd)
}
