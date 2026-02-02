package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseStorageClient implements StorageClient for Supabase Storage
type SupabaseStorageClient struct {
	supabaseURL string
	secretKey   string
	httpClient  *http.Client
}

// NewSupabaseStorageClient creates a new Supabase storage client
func NewSupabaseStorageClient(supabaseURL, secretKey string) *SupabaseStorageClient {
	return &SupabaseStorageClient{
		supabaseURL: supabaseURL,
		secretKey:   secretKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateUploadURL creates a presigned URL for uploading
func (c *SupabaseStorageClient) GenerateUploadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/upload/sign/%s/%s", c.supabaseURL, bucket, key)

	reqBody := map[string]interface{}{
		"expiresIn": expiresIn,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("storage returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.SignedURL, nil
}

// GenerateDownloadURL creates a presigned URL for downloading
func (c *SupabaseStorageClient) GenerateDownloadURL(ctx context.Context, bucket, key string, expiresIn int) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", c.supabaseURL, bucket, key)

	reqBody := map[string]interface{}{
		"expiresIn": expiresIn,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("storage returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Prepend base URL if signedURL is relative
	if len(result.SignedURL) > 0 && result.SignedURL[0] == '/' {
		return c.supabaseURL + result.SignedURL, nil
	}
	return result.SignedURL, nil
}

// Download fetches file content from storage
func (c *SupabaseStorageClient) Download(ctx context.Context, bucket, key string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.supabaseURL, bucket, key)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("storage returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return data, contentType, nil
}

// Upload stores file content to storage
func (c *SupabaseStorageClient) Upload(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.supabaseURL, bucket, key)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true") // Overwrite if exists

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Delete removes a file from storage
func (c *SupabaseStorageClient) Delete(ctx context.Context, bucket, key string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.supabaseURL, bucket, key)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 200 OK or 404 Not Found are both acceptable for delete
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
