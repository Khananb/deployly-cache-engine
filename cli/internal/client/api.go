package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Minute, // Long timeout for large uploads/downloads
		},
	}
}

type PresignResponse struct {
	URL        string `json:"url"`
	ObjectPath string `json:"object_path"`
}

type CompleteRequest struct {
	CacheKey  string `json:"cache_key"`
	SizeBytes int64  `json:"size_bytes"`
}

// RequestUploadURL asks the API for a pre-signed S3 PUT URL
func (c *Client) RequestUploadURL(cacheKey string) (*PresignResponse, error) {
	reqBody, _ := json.Marshal(map[string]string{"cache_key": cacheKey})
	
	req, err := http.NewRequest("POST", c.BaseURL+"/api/v1/cache/request-upload", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var res PresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

// RequestRestoreURL asks the API for a pre-signed S3 GET URL
func (c *Client) RequestRestoreURL(cacheKey string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/cache/restore?key=%s", c.BaseURL, cacheKey)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("cache not found")
		}
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var res PresignResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.URL, nil
}

// UploadArchive PUTs the stream directly to S3/MinIO
func (c *Client) UploadArchive(presignedURL string, data io.Reader, size int64) error {
	req, err := http.NewRequest("PUT", presignedURL, data)
	if err != nil {
		return err
	}
	req.ContentLength = size

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage upload error (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// DownloadArchive GETs the stream directly from S3/MinIO
func (c *Client) DownloadArchive(presignedURL string) (io.ReadCloser, error) {
	resp, err := c.HTTPClient.Get(presignedURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("storage download error (%d)", resp.StatusCode)
	}
	return resp.Body, nil
}

// NotifyComplete informs the API that the upload is finished and metadata should be saved
func (c *Client) NotifyComplete(cacheKey string, size int64) error {
	reqBody, _ := json.Marshal(CompleteRequest{
		CacheKey:  cacheKey,
		SizeBytes: size,
	})

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v1/cache/complete-upload", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to notify API of completion: status %d", resp.StatusCode)
	}
	return nil
}
