package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/models"
)

const (
	replicateBaseURL = "https://api.replicate.com/v1"
	hunyuan3d2Model  = "tencent/hunyuan3d-2"
)

// ReplicateAsset3DClient implements Asset3DClient using Replicate API for Hunyuan3D-2
type ReplicateAsset3DClient struct {
	apiToken   string
	storage    StorageClient
	httpClient *http.Client
}

// NewReplicateAsset3DClient creates a new 3D asset generation client
func NewReplicateAsset3DClient(apiToken string, storage StorageClient) *ReplicateAsset3DClient {
	return &ReplicateAsset3DClient{
		apiToken: apiToken,
		storage:  storage,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 3D generation can take 2-5 minutes
		},
	}
}

// Generate3DAsset creates a 3D model from an evidence photo
func (c *ReplicateAsset3DClient) Generate3DAsset(ctx context.Context, input models.Asset3DInput) (*models.Asset3DOutput, error) {
	startTime := time.Now()

	// 1. Fetch input image from storage
	imageData, mimeType, err := c.fetchImage(ctx, input.ImageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	// 2. Create prediction
	predictionID, err := c.createPrediction(ctx, imageData, mimeType, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}

	// 3. Poll for completion
	result, err := c.waitForCompletion(ctx, predictionID)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// 4. Download mesh and upload to Supabase Storage
	meshAssetKey, err := c.downloadAndUploadMesh(ctx, result.MeshURL, input)
	if err != nil {
		return nil, fmt.Errorf("failed to store mesh: %w", err)
	}

	// Determine format from output or input
	format := input.OutputFormat
	if format == "" {
		format = "glb"
	}

	return &models.Asset3DOutput{
		MeshAssetKey:   meshAssetKey,
		ThumbnailKey:   "", // TODO: Generate thumbnail from mesh
		Format:         format,
		HasTexture:     input.WithTexture,
		VertexCount:    0, // Replicate doesn't return this
		ModelUsed:      "hunyuan3d-2",
		GenerationTime: time.Since(startTime).Milliseconds(),
	}, nil
}

func (c *ReplicateAsset3DClient) fetchImage(ctx context.Context, key string) ([]byte, string, error) {
	if c.storage == nil {
		return nil, "", fmt.Errorf("storage client not configured")
	}

	bucket := "case-assets" // Default bucket name
	data, contentType, err := c.storage.Download(ctx, bucket, key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download %s: %w", key, err)
	}

	// Determine mime type
	if contentType == "" || contentType == "application/octet-stream" {
		if strings.HasSuffix(strings.ToLower(key), ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(strings.ToLower(key), ".webp") {
			contentType = "image/webp"
		} else {
			contentType = "image/jpeg"
		}
	}

	return data, contentType, nil
}

func (c *ReplicateAsset3DClient) createPrediction(ctx context.Context, imageData []byte, mimeType string, input models.Asset3DInput) (string, error) {
	// Build data URI for the image
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))

	// Build input parameters for Hunyuan3D-2
	inputParams := map[string]interface{}{
		"image":   dataURI,
		"texture": input.WithTexture,
		"seed":    42, // Fixed seed for reproducibility
	}

	// Add output format if specified
	if input.OutputFormat != "" {
		inputParams["output_format"] = input.OutputFormat
	}

	// Add prompt/description if provided
	if input.Description != "" {
		inputParams["prompt"] = input.Description
	}

	reqBody := map[string]interface{}{
		"input": inputParams,
	}

	jsonBody, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/models/%s/predictions", replicateBaseURL, hunyuan3d2Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait") // Try to get immediate response if possible

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ID, nil
}

type predictionResult struct {
	MeshURL      string
	TextureURL   string
	ThumbnailURL string
}

func (c *ReplicateAsset3DClient) waitForCompletion(ctx context.Context, predictionID string) (*predictionResult, error) {
	pollInterval := 5 * time.Second
	maxAttempts := 60 // 5 minutes max

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}

		url := fmt.Sprintf("%s/predictions/%s", replicateBaseURL, predictionID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "Bearer "+c.apiToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var prediction struct {
			Status string `json:"status"`
			Output interface{} `json:"output"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(body, &prediction); err != nil {
			continue
		}

		switch prediction.Status {
		case "succeeded":
			return c.parseOutput(prediction.Output)
		case "failed":
			return nil, fmt.Errorf("prediction failed: %s", prediction.Error)
		case "canceled":
			return nil, fmt.Errorf("prediction was canceled")
		}
		// Status is "starting" or "processing", continue polling
	}

	return nil, fmt.Errorf("prediction timed out after %d attempts", maxAttempts)
}

func (c *ReplicateAsset3DClient) parseOutput(output interface{}) (*predictionResult, error) {
	result := &predictionResult{}

	switch v := output.(type) {
	case string:
		// Single URL output (mesh file)
		result.MeshURL = v
	case map[string]interface{}:
		// Structured output with multiple files
		if mesh, ok := v["mesh"].(string); ok {
			result.MeshURL = mesh
		}
		if texture, ok := v["texture"].(string); ok {
			result.TextureURL = texture
		}
		if thumb, ok := v["thumbnail"].(string); ok {
			result.ThumbnailURL = thumb
		}
		// Try alternative keys
		if result.MeshURL == "" {
			if glb, ok := v["glb"].(string); ok {
				result.MeshURL = glb
			} else if obj, ok := v["obj"].(string); ok {
				result.MeshURL = obj
			}
		}
	case []interface{}:
		// Array output - first element is usually the mesh
		if len(v) > 0 {
			if url, ok := v[0].(string); ok {
				result.MeshURL = url
			}
		}
	}

	if result.MeshURL == "" {
		return nil, fmt.Errorf("no mesh URL in output: %v", output)
	}

	return result, nil
}

func (c *ReplicateAsset3DClient) downloadAndUploadMesh(ctx context.Context, meshURL string, input models.Asset3DInput) (string, error) {
	// Download mesh from Replicate's temporary URL
	req, err := http.NewRequestWithContext(ctx, "GET", meshURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download mesh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mesh download returned status %d", resp.StatusCode)
	}

	meshData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read mesh data: %w", err)
	}

	// Determine file extension and content type
	format := input.OutputFormat
	if format == "" {
		format = "glb"
	}

	contentType := "model/gltf-binary"
	switch format {
	case "obj":
		contentType = "text/plain"
	case "ply":
		contentType = "application/x-ply"
	case "stl":
		contentType = "application/sla"
	}

	// Generate storage key
	assetID := uuid.New().String()
	storageKey := fmt.Sprintf("cases/%s/models/%s.%s", input.CaseID, assetID, format)

	// Upload to Supabase Storage
	if c.storage != nil {
		bucket := "case-assets"
		if err := c.storage.Upload(ctx, bucket, storageKey, meshData, contentType); err != nil {
			return "", fmt.Errorf("failed to upload mesh: %w", err)
		}
	}

	return storageKey, nil
}
