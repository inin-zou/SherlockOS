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
	// Gemini model for scene analysis
	// Using gemini-2.0-flash for fast, accurate vision analysis
	// Docs: https://ai.google.dev/gemini-api/docs/models
	geminiSceneAnalysisModel = "gemini-2.0-flash-001"
)

// imagePayload holds base64 encoded image data
type imagePayload struct {
	base64Data string
	mimeType   string
}

// GeminiSceneAnalysisClient implements SceneAnalysisClient using Gemini 2.0 Flash
type GeminiSceneAnalysisClient struct {
	apiKey     string
	storage    StorageClient
	httpClient *http.Client
}

// NewGeminiSceneAnalysisClient creates a new scene analysis client
func NewGeminiSceneAnalysisClient(apiKey string, storage StorageClient) *GeminiSceneAnalysisClient {
	return &GeminiSceneAnalysisClient{
		apiKey:  apiKey,
		storage: storage,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // Vision models need longer timeout
		},
	}
}

// AnalyzeScene processes images and returns detected objects/evidence
func (c *GeminiSceneAnalysisClient) AnalyzeScene(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error) {
	startTime := time.Now()

	// 1. Fetch images from storage
	images, err := c.fetchImages(ctx, input.ImageKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch images: %w", err)
	}

	// 2. Build analysis prompt
	prompt := c.buildAnalysisPrompt(input)

	// 3. Make API request with images
	response, err := c.analyzeWithVision(ctx, images, prompt)
	if err != nil {
		return nil, fmt.Errorf("gemini vision API error: %w", err)
	}

	// 4. Parse response
	output, err := c.parseAnalysisResponse(response, input.ImageKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	output.AnalysisTime = time.Since(startTime).Milliseconds()
	output.ModelUsed = geminiSceneAnalysisModel
	return output, nil
}

func (c *GeminiSceneAnalysisClient) fetchImages(ctx context.Context, keys []string) ([]imagePayload, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage client not configured")
	}

	var images []imagePayload
	for _, key := range keys {
		// Parse bucket and path from key
		// Key format: "cases/{caseId}/scans/{batchId}/{filename}"
		bucket := "case-assets" // Default bucket name

		data, contentType, err := c.storage.Download(ctx, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("failed to download %s: %w", key, err)
		}

		// Determine mime type
		mimeType := contentType
		if mimeType == "" || mimeType == "application/octet-stream" {
			// Infer from extension
			if strings.HasSuffix(strings.ToLower(key), ".png") {
				mimeType = "image/png"
			} else if strings.HasSuffix(strings.ToLower(key), ".webp") {
				mimeType = "image/webp"
			} else {
				mimeType = "image/jpeg" // Default to JPEG
			}
		}

		images = append(images, imagePayload{
			base64Data: base64.StdEncoding.EncodeToString(data),
			mimeType:   mimeType,
		})
	}

	return images, nil
}

func (c *GeminiSceneAnalysisClient) buildAnalysisPrompt(input models.SceneAnalysisInput) string {
	basePrompt := `You are an expert forensic analyst specializing in crime scene analysis.

## Task
Analyze the provided crime scene images and:
1. Identify and catalog ALL visible objects with their positions
2. Flag any suspicious or anomalous elements
3. Identify potential evidence items
4. Describe the overall scene layout and condition

## Output Format
Respond with valid JSON in this exact structure:
{
  "objects": [
    {
      "id": "obj_001",
      "type": "furniture|door|window|wall|evidence_item|weapon|footprint|bloodstain|vehicle|person_marker|other",
      "label": "Descriptive name",
      "position_description": "Location in scene (e.g., 'center of room', 'near window')",
      "confidence": 0.0-1.0,
      "is_suspicious": true/false,
      "notes": "Any relevant observations"
    }
  ],
  "potential_evidence": ["List of potential evidence items or areas of interest"],
  "scene_description": "Overall description of the scene",
  "anomalies": ["List of anomalies or unusual observations"]
}

Be thorough and precise. Mark items as suspicious if they appear disturbed, out of place, or potentially relevant to an investigation.`

	// Add mode-specific instructions
	switch input.Mode {
	case "object_detection":
		basePrompt += "\n\n## Focus: Object Detection\nPrioritize accurate identification and cataloging of all visible objects."
	case "evidence_search":
		basePrompt += "\n\n## Focus: Evidence Search\nPrioritize identification of potential evidence and suspicious items."
	case "full_analysis":
		basePrompt += "\n\n## Focus: Full Analysis\nPerform comprehensive analysis including objects, evidence, and scene reconstruction."
	}

	// Add specific query if provided
	if input.Query != "" {
		basePrompt += fmt.Sprintf("\n\n## Specific Query\n%s", input.Query)
	}

	return basePrompt
}

func (c *GeminiSceneAnalysisClient) analyzeWithVision(ctx context.Context, images []imagePayload, prompt string) (string, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiBaseURL, geminiSceneAnalysisModel, c.apiKey)

	// Build parts array with images first, then text prompt
	var parts []map[string]interface{}

	// Add images
	for _, img := range images {
		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]interface{}{
				"mimeType": img.mimeType,
				"data":     img.base64Data,
			},
		})
	}

	// Add text prompt (must come after images)
	parts = append(parts, map[string]interface{}{
		"text": prompt,
	})

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": parts},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0.2,  // Low temperature for more deterministic output
			"topP":             0.95,
			"maxOutputTokens":  8192,
			"responseMimeType": "application/json",
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

func (c *GeminiSceneAnalysisClient) parseAnalysisResponse(response string, imageKeys []string) (*models.SceneAnalysisOutput, error) {
	jsonStr := extractJSON(response)

	var rawOutput struct {
		Objects           []models.DetectedObject `json:"objects"`
		PotentialEvidence []string                `json:"potential_evidence"`
		SceneDescription  string                  `json:"scene_description"`
		Anomalies         []string                `json:"anomalies"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawOutput); err != nil {
		// Return partial result on parse failure
		return &models.SceneAnalysisOutput{
			SceneDescription:  response[:min(500, len(response))],
			DetectedObjects:   []models.DetectedObject{},
			PotentialEvidence: []string{},
			Anomalies:         []string{},
		}, nil
	}

	// Assign IDs and source image keys to objects
	for i := range rawOutput.Objects {
		if rawOutput.Objects[i].ID == "" {
			rawOutput.Objects[i].ID = uuid.New().String()
		}
		if rawOutput.Objects[i].SourceImageKey == "" && len(imageKeys) > 0 {
			rawOutput.Objects[i].SourceImageKey = imageKeys[0]
		}
	}

	return &models.SceneAnalysisOutput{
		DetectedObjects:   rawOutput.Objects,
		PotentialEvidence: rawOutput.PotentialEvidence,
		SceneDescription:  rawOutput.SceneDescription,
		Anomalies:         rawOutput.Anomalies,
	}, nil
}
