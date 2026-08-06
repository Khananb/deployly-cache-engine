package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"deployly-cache/cli/internal/archive"
	"deployly-cache/cli/internal/client"
	"deployly-cache/cli/internal/hash"

	"github.com/spf13/cobra"
)

var (
	savePath     string
	saveLockfile string
	savePrefix   string
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Archive and upload a directory to the cache",
	Run: func(cmd *cobra.Command, args []string) {
		if apiURL == "" || apiKey == "" {
			log.Fatal("Error: DEPLOYLY_URL and DEPLOYLY_KEY must be set")
		}

		// 1. Generate Cache Key
		cacheKey, err := hash.GenerateCacheKey(savePrefix, saveLockfile)
		if err != nil {
			log.Fatalf("Failed to generate cache key: %v", err)
		}
		fmt.Printf("Generated cache key: %s\n", cacheKey)

		apiClient := client.NewClient(apiURL, apiKey)

		// 2. Check/Request Upload URL
		res, err := apiClient.RequestUploadURL(cacheKey)
		if err != nil {
			log.Fatalf("Failed to request upload URL: %v", err)
		}

		// 3. Create Archive in memory/temp
		// Note: For multi-GB files, we should use a temp file. For VPS-friendly 1GB RAM, we use a temp file.
		tempFile, err := os.CreateTemp("", "deployly-*.tar.zst")
		if err != nil {
			log.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		fmt.Printf("Archiving %s...\n", savePath)
		if err := archive.CreateArchive(savePath, tempFile); err != nil {
			log.Fatalf("Archiving failed: %v", err)
		}

		fi, _ := tempFile.Stat()
		size := fi.Size()
		fmt.Printf("Archive created: %.2f MB\n", float64(size)/(1024*1024))

		// 4. Upload directly to S3
		if _, err := tempFile.Seek(0, 0); err != nil {
			log.Fatal(err)
		}
		
		fmt.Println("Uploading to storage...")
		if err := apiClient.UploadArchive(res.URL, tempFile, size); err != nil {
			log.Fatalf("Upload failed: %v", err)
		}

		// 5. Complete Upload
		if err := apiClient.NotifyComplete(cacheKey, size); err != nil {
			log.Fatalf("Failed to finalize upload: %v", err)
		}

		fmt.Println("Cache saved successfully!")
	},
}

func init() {
	saveCmd.Flags().StringVarP(&savePath, "path", "p", "", "Directory to cache (required)")
	saveCmd.Flags().StringVarP(&saveLockfile, "lockfile", "l", "", "Lockfile to hash (e.g., go.sum) (required)")
	saveCmd.Flags().StringVar(&savePrefix, "prefix", "", "Optional key prefix")
	saveCmd.MarkFlagRequired("path")
	saveCmd.MarkFlagRequired("lockfile")
	rootCmd.AddCommand(saveCmd)
}
