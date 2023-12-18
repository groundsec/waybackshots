package cmd

import (
	"fmt"
	"os"

	"github.com/groundsec/waybackshots/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

var rootCmd = &cobra.Command{
	Use:   "waybackshots",
	Short: " Get screenshots of URLs stored in the Wayback Machine in a smart way",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("WIP!")
	},
}

func init() {
	completion := completionCmd()
	completion.Hidden = true
	rootCmd.AddCommand(completion)
	logger.SetLevel(logrus.InfoLevel)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
