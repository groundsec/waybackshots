package cmd

import (
	"fmt"
	"os"

	"github.com/groundsec/waybackshots/pkg/logger"
	"github.com/groundsec/waybackshots/pkg/screenshot"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

var (
	threads int
	url     string
	file    string
)

var rootCmd = &cobra.Command{
	Use:   "waybackshots",
	Short: "Get screenshots of URLs stored in the Wayback Machine in a smart way",
	Run: func(cmd *cobra.Command, args []string) {
		if url != "" && file != "" {
			fmt.Println("Cannot use -u/--url and -f/--file together.")
			return
		}

		if url != "" {
			screenshot.HandleUrl(url, threads)
		} else if file != "" {
			screenshot.HandleFile(file, threads)
		} else {
			fmt.Println("Please specify either -u/--url or -f/--file.")
		}
	},
}

func init() {
	completion := completionCmd()
	completion.Hidden = true
	logger.SetLevel(logrus.InfoLevel)
	rootCmd.AddCommand(completion)
	rootCmd.PersistentFlags().IntVarP(&threads, "threads", "t", 5, "number of workers to use")
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "URL to analyze")
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "file to read")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
