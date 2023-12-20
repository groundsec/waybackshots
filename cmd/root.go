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
	url       string
	file      string
	outputDir string
	verbose   bool
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
			screenshot.HandleUrl(url, outputDir)
		} else if file != "" {
			screenshot.HandleFile(file, outputDir)
		} else {
			fmt.Println("Please specify either -u/--url or -f/--file.")
		}
	},
}

func init() {
	completion := completionCmd()
	completion.Hidden = true
	rootCmd.AddCommand(completion)
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "URL to screenshot")
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "File with URLs to screenshot")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", ".", "Output dir")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	if verbose {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.ErrorLevel)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
