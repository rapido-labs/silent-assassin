package cmd

import (
	"fmt"
	"os"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/spf13/cobra"
)

var cfgFile string
var configProvider *config.Provider
var zapLogger logger.ZapLogger

var rootCmd = &cobra.Command{
	Use:   "silent-assassin",
	Short: "Preemptible Node Killer",
	Long:  `Handles graceful shutdown and termination of preemptible nodes.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/.silent-assassin.yaml", "config file (default is $HOME/.silent-assassin.yaml)")
}

func initConfig() {

}
