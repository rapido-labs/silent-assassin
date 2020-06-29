package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
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

		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		wg := &sync.WaitGroup{}
		ctx, cancelFn := context.WithCancel(context.Background())

		configProvider := config.Init(cfgFile)
		zapLogger := logger.Init(configProvider)
		kubeClient := k8s.NewClient(configProvider, zapLogger)
		gcloudClient, err := gcloud.NewClient(zapLogger)
		if err != nil {
			zapLogger.Error("Error creating gcloud client")
		}

		ss := spotter.NewSpotterService(configProvider, zapLogger, kubeClient)
		wg.Add(1)
		go ss.Start(ctx, wg)

		ks := killer.NewKillerService(configProvider, zapLogger, kubeClient, gcloudClient)
		wg.Add(1)
		go ks.Start(ctx, wg)

		<-sigChan

		zapLogger.Info("Starting shut down")
		cancelFn()
		wg.Wait()
		zapLogger.Info("Shut down completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
