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
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

	// Reasoning model - Gemini 2.5 Flash with thinking capability
	// Docs: https://ai.google.dev/gemini-api/docs/thinking
	geminiReasoningModel = "gemini-2.5-flash"

	// Profile extraction model - also uses Gemini 2.5 Flash
	geminiProfileModel = "gemini-2.5-flash"

	// Image generation models (Nano Banana series)
	// Docs: https://ai.google.dev/gemini-api/docs/image-generation
	geminiImageModel1K = "gemini-2.5-flash-image"     // Nano Banana - fast, low cost (~$0.04/image)
	geminiImageModelHQ = "gemini-3-pro-image-preview" // Nano Banana Pro - high quality, 2K/4K (~$0.134-0.24/image)
)

// GeminiClient implements AI clients using Google's Gemini API
type GeminiClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ============================================
// REASONING CLIENT IMPLEMENTATION
// ============================================

// GeminiReasoningClient implements ReasoningClient using Gemini 2.5 Flash with Thinking
type GeminiReasoningClient struct {
	*GeminiClient
}

// NewGeminiReasoningClient creates a new reasoning client
func NewGeminiReasoningClient(apiKey string) *GeminiReasoningClient {
	return &GeminiReasoningClient{
		GeminiClient: NewGeminiClient(apiKey),
	}
}

// Reason generates trajectory hypotheses using Gemini
func (c *GeminiReasoningClient) Reason(ctx context.Context, input models.ReasoningInput) (*models.ReasoningOutput, error) {
	startTime := time.Now()

	// Build prompt
	prompt := c.buildReasoningPrompt(input)

	// Make API request using Gemini 2.5 Flash with thinking
	response, err := c.generateContent(ctx, prompt, input.ThinkingBudget, geminiReasoningModel)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	// Parse response into trajectories
	output, err := c.parseReasoningResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	output.ModelStats.LatencyMs = time.Since(startTime).Milliseconds()
	return output, nil
}

func (c *GeminiReasoningClient) buildReasoningPrompt(input models.ReasoningInput) string {
	sgJSON, _ := json.MarshalIndent(input.Scenegraph, "", "  ")

	prompt := fmt.Sprintf(`You are an expert forensic analyst. Analyze the following crime scene data and generate plausible movement trajectory hypotheses.

## Scene Graph (World State)
%s

## Task
1. Analyze the positions of objects, evidence, and constraints
2. Generate %d ranked trajectory hypotheses explaining possible suspect movements
3. For each trajectory, provide:
   - A sequence of movement segments with positions
   - Confidence score (0-1) for each segment
   - Explanation referencing specific evidence
   - Evidence references that support or contradict the hypothesis

## Output Format
Respond with valid JSON in this exact structure:
{
  "trajectories": [
    {
      "id": "traj_001",
      "rank": 1,
      "overall_confidence": 0.75,
      "segments": [
        {
          "id": "seg_001",
          "from_position": [x, y, z],
          "to_position": [x, y, z],
          "confidence": 0.8,
          "explanation": "Entry through main door based on...",
          "evidence_refs": [
            {"evidence_id": "ev_001", "relevance": "supports", "weight": 0.9}
          ]
        }
      ]
    }
  ],
  "uncertainty_areas": [],
  "next_step_suggestions": [
    {"type": "collect_evidence", "description": "...", "priority": "high"}
  ],
  "thinking_summary": "Brief summary of reasoning process"
}`, string(sgJSON), input.MaxTrajectories)

	// Add constraint overrides if present
	if len(input.ConstraintsOverride) > 0 {
		constraintsJSON, _ := json.Marshal(input.ConstraintsOverride)
		prompt += fmt.Sprintf("\n\n## Additional Constraints\n%s", string(constraintsJSON))
	}

	return prompt
}

func (c *GeminiReasoningClient) parseReasoningResponse(response string) (*models.ReasoningOutput, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := extractJSON(response)

	var output models.ReasoningOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		// If parsing fails, create a minimal valid response
		output = models.ReasoningOutput{
			Trajectories: []models.Trajectory{
				{
					ID:                uuid.New().String(),
					Rank:              1,
					OverallConfidence: 0.5,
					Segments:          []models.TrajectorySegment{},
				},
			},
			ThinkingSummary: "Response parsing incomplete: " + response[:min(200, len(response))],
			ModelStats:      models.ModelStats{},
		}
	}

	return &output, nil
}

// ============================================
// PROFILE CLIENT IMPLEMENTATION
// ============================================

// GeminiProfileClient implements ProfileClient using Gemini
type GeminiProfileClient struct {
	*GeminiClient
}

// NewGeminiProfileClient creates a new profile extraction client
func NewGeminiProfileClient(apiKey string) *GeminiProfileClient {
	return &GeminiProfileClient{
		GeminiClient: NewGeminiClient(apiKey),
	}
}

// ExtractProfile extracts suspect attributes from witness statements
func (c *GeminiProfileClient) ExtractProfile(ctx context.Context, statements []models.WitnessStatementInput, existing *models.SuspectAttributes) (*models.SuspectAttributes, error) {
	prompt := c.buildProfilePrompt(statements, existing)

	// Use profile model for extraction
	response, err := c.generateContent(ctx, prompt, 4096, geminiProfileModel)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	return c.parseProfileResponse(response, statements)
}

func (c *GeminiProfileClient) buildProfilePrompt(statements []models.WitnessStatementInput, existing *models.SuspectAttributes) string {
	var stmtTexts []string
	for i, stmt := range statements {
		stmtTexts = append(stmtTexts, fmt.Sprintf("Statement %d (Source: %s, Credibility: %.2f):\n%s",
			i+1, stmt.SourceName, stmt.Credibility, stmt.Content))
	}

	existingJSON := "{}"
	if existing != nil {
		existingBytes, _ := json.Marshal(existing)
		existingJSON = string(existingBytes)
	}

	return fmt.Sprintf(`You are an expert at extracting physical descriptions from witness statements.

## Witness Statements
%s

## Existing Profile (if any)
%s

## Task
Extract structured physical attributes from the statements. Weight each extraction by the source credibility.

## Output Format
Respond with valid JSON:
{
  "age_range": {"min": 25, "max": 35, "confidence": 0.7},
  "height_range_cm": {"min": 170, "max": 180, "confidence": 0.8},
  "build": {"value": "athletic", "confidence": 0.6},
  "hair_color": {"value": "dark brown", "confidence": 0.7},
  "hair_style": {"value": "short", "confidence": 0.5},
  "eye_color": {"value": "brown", "confidence": 0.4},
  "skin_tone": {"value": "medium", "confidence": 0.6},
  "facial_hair": {"value": "stubble", "confidence": 0.5},
  "distinguishing_features": [{"value": "scar on left cheek", "confidence": 0.8}],
  "clothing_description": {"value": "dark hoodie, jeans", "confidence": 0.7},
  "conflicts": [
    {"attribute": "hair_color", "values": ["brown", "black"], "source_indices": [0, 1]}
  ]
}

Only include attributes that are mentioned in the statements.`, strings.Join(stmtTexts, "\n\n"), string(existingJSON))
}

func (c *GeminiProfileClient) parseProfileResponse(response string, statements []models.WitnessStatementInput) (*models.SuspectAttributes, error) {
	jsonStr := extractJSON(response)

	var parsed struct {
		AgeRange            *models.RangeAttribute  `json:"age_range"`
		HeightRangeCm       *models.RangeAttribute  `json:"height_range_cm"`
		Build               *models.StringAttribute `json:"build"`
		HairColor           string                  `json:"hair_color"`
		HairStyle           string                  `json:"hair_style"`
		SkinTone            *models.StringAttribute `json:"skin_tone"`
		FacialHair          *models.StringAttribute `json:"facial_hair"`
		Glasses             *models.StringAttribute `json:"glasses"`
		DistinctiveFeatures []struct {
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
		} `json:"distinctive_features"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Return empty attributes on parse failure
		return models.NewEmptySuspectAttributes(), nil
	}

	attrs := models.NewEmptySuspectAttributes()
	attrs.AgeRange = parsed.AgeRange
	attrs.HeightRangeCm = parsed.HeightRangeCm
	attrs.Build = parsed.Build
	attrs.SkinTone = parsed.SkinTone
	attrs.FacialHair = parsed.FacialHair
	attrs.Glasses = parsed.Glasses

	// Combine hair color and style into HairAttribute
	if parsed.HairColor != "" || parsed.HairStyle != "" {
		attrs.Hair = &models.HairAttribute{
			Color:      parsed.HairColor,
			Style:      parsed.HairStyle,
			Confidence: 0.7, // Default confidence
		}
	}

	// Convert distinctive features
	for _, f := range parsed.DistinctiveFeatures {
		attrs.DistinctiveFeatures = append(attrs.DistinctiveFeatures, models.FeatureAttribute{
			Description: f.Description,
			Confidence:  f.Confidence,
		})
	}

	return attrs, nil
}

// ============================================
// IMAGE GENERATION CLIENT IMPLEMENTATION
// ============================================

// GeminiImageGenClient implements ImageGenClient using Gemini
type GeminiImageGenClient struct {
	*GeminiClient
}

// NewGeminiImageGenClient creates a new image generation client
func NewGeminiImageGenClient(apiKey string) *GeminiImageGenClient {
	return &GeminiImageGenClient{
		GeminiClient: NewGeminiClient(apiKey),
	}
}

// Generate creates an image based on the input parameters
func (c *GeminiImageGenClient) Generate(ctx context.Context, input models.ImageGenInput) (*models.ImageGenOutput, error) {
	startTime := time.Now()

	// Handle POV generation specially (multiple images)
	if input.GenType == models.ImageGenTypeScenePOV {
		return c.generatePOVImages(ctx, input, startTime)
	}

	prompt := c.buildImagePrompt(input)
	model := input.GetModelForResolution()

	// Make image generation request
	imageData, err := c.generateImage(ctx, prompt, model)
	if err != nil {
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	// Determine dimensions based on resolution
	width, height := 1024, 1024
	cost := 0.04
	if input.Resolution == "2k" {
		width, height = 2048, 2048
		cost = 0.134
	} else if input.Resolution == "4k" {
		width, height = 4096, 4096
		cost = 0.24
	}

	// Generate asset key
	assetID := uuid.New().String()
	assetKey := fmt.Sprintf("cases/%s/generated/%s.png", input.CaseID, assetID)
	thumbnailKey := fmt.Sprintf("cases/%s/generated/%s_thumb.png", input.CaseID, assetID)

	// In production, upload imageData to Supabase Storage here
	_ = imageData

	return &models.ImageGenOutput{
		AssetKey:       assetKey,
		ThumbnailKey:   thumbnailKey,
		Width:          width,
		Height:         height,
		ModelUsed:      model,
		GenerationTime: time.Since(startTime).Milliseconds(),
		CostUSD:        cost,
	}, nil
}

// generatePOVImages generates multiple consistent POV images for scene reconstruction
func (c *GeminiImageGenClient) generatePOVImages(ctx context.Context, input models.ImageGenInput, startTime time.Time) (*models.ImageGenOutput, error) {
	model := input.GetModelForResolution()

	// Determine dimensions based on resolution
	width, height := 1024, 1024
	costPerImage := 0.04
	if input.Resolution == "2k" {
		width, height = 2048, 2048
		costPerImage = 0.134
	} else if input.Resolution == "4k" {
		width, height = 4096, 4096
		costPerImage = 0.24
	}

	var generatedImages []models.GeneratedImage
	var totalCost float64

	// Generate an image for each view angle
	for _, viewAngle := range input.ViewAngles {
		// Create a copy of input with the specific view angle
		povInput := input
		povInput.StylePrompt = viewAngle

		prompt := c.buildScenePOVPrompt(povInput)

		// Generate the image
		imageData, err := c.generateImage(ctx, prompt, model)
		if err != nil {
			// Log error but continue with other views
			fmt.Printf("Warning: Failed to generate %s view: %v\n", viewAngle, err)
			continue
		}

		// Generate asset keys
		assetID := uuid.New().String()
		assetKey := fmt.Sprintf("cases/%s/generated/pov/%s_%s.png", input.CaseID, viewAngle, assetID)
		thumbnailKey := fmt.Sprintf("cases/%s/generated/pov/%s_%s_thumb.png", input.CaseID, viewAngle, assetID)

		// In production, upload imageData to Supabase Storage here
		_ = imageData

		generatedImages = append(generatedImages, models.GeneratedImage{
			ViewAngle:    viewAngle,
			AssetKey:     assetKey,
			ThumbnailKey: thumbnailKey,
			Width:        width,
			Height:       height,
		})

		totalCost += costPerImage
	}

	if len(generatedImages) == 0 {
		return nil, fmt.Errorf("failed to generate any POV images")
	}

	// Return the first image as the primary output, with all images in GeneratedImages
	return &models.ImageGenOutput{
		AssetKey:        generatedImages[0].AssetKey,
		ThumbnailKey:    generatedImages[0].ThumbnailKey,
		Width:           width,
		Height:          height,
		ModelUsed:       model,
		GenerationTime:  time.Since(startTime).Milliseconds(),
		CostUSD:         totalCost,
		GeneratedImages: generatedImages,
	}, nil
}

func (c *GeminiImageGenClient) buildImagePrompt(input models.ImageGenInput) string {
	switch input.GenType {
	case models.ImageGenTypePortrait:
		return c.buildPortraitPrompt(input.PortraitAttrs, input.StylePrompt)
	case models.ImageGenTypeEvidenceBoard:
		return fmt.Sprintf("Generate an evidence board visualization showing connected clues and evidence items. Style: %s", input.StylePrompt)
	case models.ImageGenTypeScenePOV:
		return c.buildScenePOVPrompt(input)
	case models.ImageGenTypeAssetClean:
		return c.buildAssetCleanPrompt(input)
	default:
		return input.StylePrompt
	}
}

// buildScenePOVPrompt creates a prompt for generating consistent POV images
func (c *GeminiImageGenClient) buildScenePOVPrompt(input models.ImageGenInput) string {
	roomType := input.RoomType
	if roomType == "" {
		roomType = "indoor room"
	}

	// This will be called multiple times, once per view angle
	// The caller should set StylePrompt to the specific view angle
	viewAngle := input.StylePrompt
	if viewAngle == "" {
		viewAngle = "front view"
	}

	prompt := fmt.Sprintf(`Generate a photorealistic image of a %s from a %s perspective.

## Scene Description
%s

## Requirements
- Consistent with other views of the same space
- Professional forensic documentation style photography
- Uniform, well-balanced lighting (as if using crime scene lights)
- Sharp details, no motion blur
- Neutral color temperature
- High detail on all surfaces and objects
- The perspective should clearly show the spatial relationships

## Camera Position: %s
- If "front": Looking at the main focal point of the room
- If "left": Looking from the left side towards the right
- If "right": Looking from the right side towards the left
- If "back": Looking from the back towards the entrance
- If "corner_nw": Looking from the northwest corner diagonally
- If "corner_se": Looking from the southeast corner diagonally

Generate a single, high-quality image showing this exact view.`, roomType, viewAngle, input.SceneDescription, viewAngle)

	return prompt
}

// buildAssetCleanPrompt creates a prompt for generating clean isolated object images
func (c *GeminiImageGenClient) buildAssetCleanPrompt(input models.ImageGenInput) string {
	return fmt.Sprintf(`Generate a forensic evidence photograph of the following object:

## Object Description
%s

## Requirements
- Isolated on a pure white background
- Studio lighting with soft shadows
- Multiple angles visible if possible (or clear single angle)
- Sharp focus, high detail
- Scale reference implied by composition
- Professional evidence documentation style
- No other objects in frame
- Clean, clinical presentation

The image should be suitable for 3D model generation.`, input.ObjectDescription)
}

func (c *GeminiImageGenClient) buildPortraitPrompt(attrs *models.SuspectAttributes, stylePrompt string) string {
	parts := []string{"Generate a realistic portrait of a person with the following characteristics:"}

	if attrs != nil {
		if attrs.AgeRange != nil {
			parts = append(parts, fmt.Sprintf("- Age: approximately %d-%d years old", int(attrs.AgeRange.Min), int(attrs.AgeRange.Max)))
		}
		if attrs.Build != nil {
			parts = append(parts, fmt.Sprintf("- Build: %s", attrs.Build.Value))
		}
		if attrs.Hair != nil {
			if attrs.Hair.Color != "" {
				parts = append(parts, fmt.Sprintf("- Hair color: %s", attrs.Hair.Color))
			}
			if attrs.Hair.Style != "" {
				parts = append(parts, fmt.Sprintf("- Hair style: %s", attrs.Hair.Style))
			}
		}
		if attrs.SkinTone != nil {
			parts = append(parts, fmt.Sprintf("- Skin tone: %s", attrs.SkinTone.Value))
		}
		if attrs.FacialHair != nil {
			parts = append(parts, fmt.Sprintf("- Facial hair: %s", attrs.FacialHair.Value))
		}
		if attrs.Glasses != nil {
			parts = append(parts, fmt.Sprintf("- Glasses: %s", attrs.Glasses.Value))
		}
		for _, feature := range attrs.DistinctiveFeatures {
			parts = append(parts, fmt.Sprintf("- Distinguishing feature: %s", feature.Description))
		}
	}

	if stylePrompt != "" {
		parts = append(parts, fmt.Sprintf("\nStyle: %s", stylePrompt))
	} else {
		parts = append(parts, "\nStyle: Photorealistic, neutral expression, front-facing, studio lighting")
	}

	return strings.Join(parts, "\n")
}

func (c *GeminiImageGenClient) generateImage(ctx context.Context, prompt, model string) ([]byte, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiBaseURL, model, c.apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"IMAGE", "TEXT"},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to extract image data
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find image data in response
	for _, candidate := range result.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MimeType, "image/") {
				return base64.StdEncoding.DecodeString(part.InlineData.Data)
			}
		}
	}

	return nil, fmt.Errorf("no image data in response")
}

// ============================================
// SHARED HELPERS
// ============================================

func (c *GeminiClient) generateContent(ctx context.Context, prompt string, thinkingBudget int, model string) (string, error) {
	if model == "" {
		model = geminiReasoningModel
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiBaseURL, model, c.apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"topP":            0.95,
			"maxOutputTokens": 8192,
		},
	}

	// Add thinking config if budget specified
	if thinkingBudget > 0 {
		reqBody["generationConfig"].(map[string]interface{})["thinkingConfig"] = map[string]interface{}{
			"thinkingBudget": thinkingBudget,
		}
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
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

// extractJSON extracts JSON from a response that may contain markdown code blocks
func extractJSON(response string) string {
	// Try to find JSON in code block
	if start := strings.Index(response, "```json"); start != -1 {
		start += 7
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	// Try to find JSON in generic code block
	if start := strings.Index(response, "```"); start != -1 {
		start += 3
		// Skip language identifier if present
		if newline := strings.Index(response[start:], "\n"); newline != -1 {
			start += newline + 1
		}
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	// Try to find JSON object directly
	if start := strings.Index(response, "{"); start != -1 {
		depth := 0
		for i := start; i < len(response); i++ {
			if response[i] == '{' {
				depth++
			} else if response[i] == '}' {
				depth--
				if depth == 0 {
					return response[start : i+1]
				}
			}
		}
	}

	return response
}

// GetModelForResolution returns the appropriate Nano Banana model based on resolution
func GetModelForResolution(resolution string) string {
	if resolution == "2k" || resolution == "4k" {
		return geminiImageModelHQ // Nano Banana Pro for high quality
	}
	return geminiImageModel1K // Nano Banana for fast iteration
}
