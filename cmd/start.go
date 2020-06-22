package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
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

		ss := spotter.NewSpotterService(configProvider, zapLogger, kubeClient)
		wg.Add(1)
		go ss.Start(ctx, wg)

		<-sigChan
		cancelFn()
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
