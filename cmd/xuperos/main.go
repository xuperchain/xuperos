package main

import (
	"log"

	"github.com/xuperchain/xuperos/cmd/xuperos/cmd"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd, err := NewServiceCommand()
	if err != nil {
		log.Fatalf("start service failed.err:%v", err)
	}

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("start service failed.err:%v", err)
	}
}

func NewServiceCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:           "xuperos <command> [arguments]",
		Short:         "xuperos is a blockchain network building service.",
		Long:          "xuperos is a blockchain network building service.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       "xuperos startup --conf /home/rd/xuperos/conf/env.yaml",
	}

	// cmd service
	rootCmd.AddCommand(cmd.GetStartupCmd().GetCmd())
	// cmd version
	rootCmd.AddCommand(cmd.GetVersionCmd().GetCmd())

	return rootCmd, nil
}
