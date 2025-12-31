package cmd

import (
	"fmt"
	"os"

	"drcom-go/pkg/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drcom",
	Short: "Dr.COM Client for Linux/Headless",
	Long:  `A CLI tool to manage Dr.COM network login/logout and monitoring.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)
}
