package main

import (
	"github.com/buni/wallet/cmd/api"
	"github.com/buni/wallet/cmd/worker"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	root := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet API Service",
		Long:  "Wallet API Service",
	}

	root.AddCommand(api.NewCommand())
	root.AddCommand(worker.NewCommand())

	if err := root.Execute(); err != nil {
		zap.L().Sugar().Fatalln("failed to execute command", err)
	}
}
