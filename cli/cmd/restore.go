package cmd

import (
	"fmt"
	"log"

	"deployly-cache/cli/internal/archive"
	"deployly-cache/cli/internal/client"
	"deployly-cache/cli/internal/hash"

	"github.com/spf13/cobra"
)

var (
	restorePath     string
	restoreLockfile string
	restorePrefix   string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Download and extract a directory from the cache",
	Run: func(cmd *cobra.Command, args []string) {
		if apiURL == "" || apiKey == "" {
			log.Fatal("Error: DEPLOYLY_URL and DEPLOYLY_KEY must be set")
		}

		// 1. Generate Cache Key
		cacheKey, err := hash.GenerateCacheKey(restorePrefix, restoreLockfile)
		if err != nil {
			log.Fatalf("Failed to generate cache key: %v", err)
		}
		fmt.Printf("Looking for cache key: %s\n", cacheKey)

		apiClient := client.NewClient(apiURL, apiKey)

		// 2. Request Download URL
		downloadURL, err := apiClient.RequestRestoreURL(cacheKey)
		if err != nil {
			if err.Error() == "cache not found" {
				fmt.Println("Cache miss. Proceeding without restore.")
				return
			}
			log.Fatalf("Failed to request restore URL: %v", err)
		}

		// 3. Stream Download and Extract
		fmt.Println("Cache hit! Downloading and extracting...")
		
		stream, err := apiClient.DownloadArchive(downloadURL)
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}
		defer stream.Close()

		if err := archive.ExtractArchive(stream, restorePath); err != nil {
			log.Fatalf("Extraction failed: %v", err)
		}

		fmt.Println("Cache restored successfully!")
	},
}

func init() {
	restoreCmd.Flags().StringVarP(&restorePath, "path", "p", ".", "Target directory for restoration")
	restoreCmd.Flags().StringVarP(&restoreLockfile, "lockfile", "l", "", "Lockfile to hash (required)")
	restoreCmd.Flags().StringVar(&restorePrefix, "prefix", "", "Optional key prefix")
	restoreCmd.MarkFlagRequired("lockfile")
	rootCmd.AddCommand(restoreCmd)
}
