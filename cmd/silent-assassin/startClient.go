package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/client"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/spf13/cobra"
)

var startClientCmd = &cobra.Command{
	Use:   "startClient",
	Short: "starts the Silent Assassin client",
	Long: `startClient starts the Silent Assassin client which runs on a
	kubernetes node and notifies Silent Assassin server when it recieves preemption
	notification`,
	Run: func(cmd *cobra.Command, args []string) {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		wg := &sync.WaitGroup{}
		ctx, cancelFn := context.WithCancel(context.Background())

		configProvider := config.Init(cfgFile)
		zapLogger := logger.Init(configProvider)

		pns := client.NewPreemptionNotificationService(zapLogger, configProvider)
		wg.Add(1)
		go pns.Start(ctx, wg)
		<-sigChan

		zapLogger.Info("Starting shut down")
		cancelFn()
		wg.Wait()
		zapLogger.Info("Shut down completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(startClientCmd)
}
