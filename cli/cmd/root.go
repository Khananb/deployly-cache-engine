package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string
)

var rootCmd = &cobra.Command{
	Use:   "deployly",
	Short: "Deployly Cache CLI - High performance CI/CD caching",
	Long:  `A universal dependency caching tool for CI/CD pipelines optimized for Go, Node.js, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "url", os.Getenv("DEPLOYLY_URL"), "Deployly API URL (env: DEPLOYLY_URL)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "key", os.Getenv("DEPLOYLY_KEY"), "Deployly API Key (env: DEPLOYLY_KEY)")
}
