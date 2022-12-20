package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/httpserver"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/roppenlabs/silent-assassin/pkg/shifter"
	"github.com/roppenlabs/silent-assassin/pkg/spotter"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts silent assassin server",

	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		wg := &sync.WaitGroup{}
		ctx, cancelFn := context.WithCancel(context.Background())

		configProvider := config.Init(cfgFile)
		zapLogger := logger.Init(configProvider)
		kubeClient := k8s.NewClient(configProvider, zapLogger)

		gcloudClient := gcloud.NewClient(kubeClient)

		ns := notifier.NewNotificationService(configProvider, zapLogger)
		wg.Add(1)
		go ns.Start(ctx, wg)

		ss := spotter.NewSpotterService(configProvider, zapLogger, kubeClient, ns)
		wg.Add(1)
		go ss.Start(ctx, wg)

		ks := killer.NewKillerService(configProvider, zapLogger, kubeClient, gcloudClient, ns)
		wg.Add(1)
		go ks.Start(ctx, wg)

		if configProvider.GetBool(config.ShifterEnabled) {
			shs := shifter.NewShifterService(configProvider, zapLogger, kubeClient, gcloudClient, ns, ks)
			wg.Add(1)
			go shs.Start(ctx, wg)
		}

		server := httpserver.New(configProvider, zapLogger, ks)
		wg.Add(1)
		go server.Start(ctx, wg)

		<-sigChan

		zapLogger.Info("Starting shut down")
		cancelFn()
		wg.Wait()
		zapLogger.Info("Shut down completed successfully")
	},
}

func init() {
	startCmd.AddCommand(serverCmd)
}
