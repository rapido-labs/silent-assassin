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
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/roppenlabs/silent-assassin/pkg/client"
	"github.com/roppenlabs/silent-assassin/pkg/config"
	"github.com/roppenlabs/silent-assassin/pkg/logger"
	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "starts silent-assassin client",
	// Long description will be updated once entire refactoring is done.
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		wg := &sync.WaitGroup{}
		ctx, cancelFn := context.WithCancel(context.Background())

		configProvider := config.Init(cfgFile)
		zapLogger := logger.Init(configProvider)

		pns := client.NewPreemptionNotifier(zapLogger, configProvider)
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
	startCmd.AddCommand(clientCmd)
}
