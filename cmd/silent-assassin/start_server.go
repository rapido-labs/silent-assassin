/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/gcloud"
	"github.com/roppenlabs/silent-assassin/pkg/k8s"
	"github.com/roppenlabs/silent-assassin/pkg/killer"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/roppenlabs/silent-assassin/pkg/notifier"
	"github.com/roppenlabs/silent-assassin/pkg/server"
	"github.com/roppenlabs/silent-assassin/pkg/spotter"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts silent assassin server",
	// Long description will be updated once entire refactoring is done.
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		wg := &sync.WaitGroup{}
		ctx, cancelFn := context.WithCancel(context.Background())

		configProvider := config.Init(cfgFile)
		zapLogger := logger.Init(configProvider)
		kubeClient := k8s.NewClient(configProvider, zapLogger)
		gcloudClient, err := gcloud.NewClient()
		if err != nil {
			zapLogger.Error(fmt.Sprintf("Error creating gcloud client: %s", err.Error()))
		}

		ns := notifier.NewNotifier(configProvider, zapLogger)
		wg.Add(1)
		go ns.Start(ctx, wg)

		ss := spotter.NewSpotterService(configProvider, zapLogger, kubeClient, ns)
		wg.Add(1)
		go ss.Start(ctx, wg)

		ks := killer.NewKillerService(configProvider, zapLogger, kubeClient, gcloudClient, ns)
		wg.Add(1)
		go ks.Start(ctx, wg)

		server := server.NewServer(configProvider, zapLogger, ks)
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
