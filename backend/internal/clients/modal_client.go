package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/models"
)

// ============================================
// MODAL RECONSTRUCTION CLIENT (HunyuanWorld-Mirror)
// ============================================

// ModalReconstructionClient implements ReconstructionClient using Modal-hosted HunyuanWorld-Mirror
type ModalReconstructionClient struct {
	baseURL    string
	storage    StorageClient
	httpClient *http.Client
}

// NewModalReconstructionClient creates a new Modal reconstruction client
// baseURL should be like "https://ykzou1214--sherlock-mirror"
func NewModalReconstructionClient(baseURL string, storage StorageClient) *ModalReconstructionClient {
	return &ModalReconstructionClient{
		baseURL: baseURL,
		storage: storage,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // Reconstruction can take time
		},
	}
}

// modalReconstructionRequest is the request format for the Modal API
type modalReconstructionRequest struct {
	CaseID             string      `json:"case_id"`
	ScanAssetKeys      []string    `json:"scan_asset_keys"` // Base64 encoded images
	CameraPoses        interface{} `json:"camera_poses,omitempty"`
	ExistingScenegraph interface{} `json:"existing_scenegraph,omitempty"`
}

// modalReconstructionResponse is the response format from the Modal API
type modalReconstructionResponse struct {
	Objects []struct {
		ID         string  `json:"id"`
		Action     string  `json:"action"`
		Confidence float64 `json:"confidence"`
		Object     struct {
			ID         string                 `json:"id"`
			Type       string                 `json:"type"`
			Label      string                 `json:"label"`
			Pose       map[string]interface{} `json:"pose"`
			BBox       map[string]interface{} `json:"bbox"`
			State      string                 `json:"state"`
			Confidence float64                `json:"confidence"`
		} `json:"object"`
		SourceImages []string `json:"source_images"`
	} `json:"objects"`
	MeshAssetKey       *string `json:"mesh_asset_key"`
	PointcloudAssetKey *string `json:"pointcloud_asset_key"`
	UncertaintyRegions []struct {
		ID     string                 `json:"id"`
		BBox   map[string]interface{} `json:"bbox"`
		Level  string                 `json:"level"`
		Reason string                 `json:"reason"`
	} `json:"uncertainty_regions"`
	ProcessingStats struct {
		InputImages      int   `json:"input_images"`
		DetectedObjects  int   `json:"detected_objects"`
		ProcessingTimeMs int64 `json:"processing_time_ms"`
	} `json:"processing_stats"`
}

// Reconstruct processes scan images and returns scene updates using Modal HunyuanWorld-Mirror
func (c *ModalReconstructionClient) Reconstruct(ctx context.Context, input models.ReconstructionInput) (*models.ReconstructionOutput, error) {
	// 1. Fetch and encode images from storage
	encodedImages, err := c.fetchAndEncodeImages(ctx, input.ScanAssetKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch images: %w", err)
	}

	// 2. Build request
	reqBody := modalReconstructionRequest{
		CaseID:        input.CaseID,
		ScanAssetKeys: encodedImages,
	}

	if input.ExistingScenegraph != nil {
		reqBody.ExistingScenegraph = input.ExistingScenegraph
	}

	// 3. Make API call
	jsonBody, _ := json.Marshal(reqBody)
	url := c.baseURL + "-reconstruct.modal.run"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 4. Parse response
	var modalResp modalReconstructionResponse
	if err := json.Unmarshal(body, &modalResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 5. Convert to models.ReconstructionOutput
	return c.convertToOutput(modalResp, input.ScanAssetKeys), nil
}

func (c *ModalReconstructionClient) fetchAndEncodeImages(ctx context.Context, keys []string) ([]string, error) {
	if c.storage == nil {
		// If no storage client, assume keys are already base64 encoded
		return keys, nil
	}

	var encoded []string
	bucket := "assets"

	for _, key := range keys {
		data, _, err := c.storage.Download(ctx, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("failed to download %s: %w", key, err)
		}
		encoded = append(encoded, base64.StdEncoding.EncodeToString(data))
	}

	return encoded, nil
}

func (c *ModalReconstructionClient) convertToOutput(resp modalReconstructionResponse, sourceKeys []string) *models.ReconstructionOutput {
	output := &models.ReconstructionOutput{
		Objects:            []models.SceneObjectProposal{},
		UncertaintyRegions: []models.UncertaintyRegion{},
		ProcessingStats: models.ProcessingStats{
			InputImages:      resp.ProcessingStats.InputImages,
			DetectedObjects:  resp.ProcessingStats.DetectedObjects,
			ProcessingTimeMs: resp.ProcessingStats.ProcessingTimeMs,
		},
	}

	// Convert objects
	for _, obj := range resp.Objects {
		proposal := models.SceneObjectProposal{
			ID:           obj.ID,
			Action:       obj.Action,
			Confidence:   obj.Confidence,
			SourceImages: obj.SourceImages,
		}

		if obj.Object.ID != "" {
			sceneObj := &models.SceneObject{
				ID:         obj.Object.ID,
				Type:       models.ObjectType(obj.Object.Type),
				Label:      obj.Object.Label,
				State:      models.ObjectState(obj.Object.State),
				Confidence: obj.Object.Confidence,
			}

			// Parse pose
			if pose := obj.Object.Pose; pose != nil {
				sceneObj.Pose = parsePose(pose)
			}

			// Parse bounding box
			if bbox := obj.Object.BBox; bbox != nil {
				sceneObj.BBox = parseBBox(bbox)
			}

			proposal.Object = sceneObj
		}

		output.Objects = append(output.Objects, proposal)
	}

	// Convert uncertainty regions
	for _, region := range resp.UncertaintyRegions {
		ur := models.UncertaintyRegion{
			ID:     region.ID,
			Level:  models.UncertaintyLevel(region.Level),
			Reason: region.Reason,
		}

		if bbox := region.BBox; bbox != nil {
			ur.BBox = parseBBox(bbox)
		}

		output.UncertaintyRegions = append(output.UncertaintyRegions, ur)
	}

	return output
}

// ============================================
// MODAL REPLAY CLIENT (HY-World-1.5 / WorldPlay)
// ============================================

// ModalReplayClient implements ReplayClient using Modal-hosted HY-WorldPlay
type ModalReplayClient struct {
	baseURL    string
	storage    StorageClient
	httpClient *http.Client
}

// NewModalReplayClient creates a new Modal replay client
// baseURL should be like "https://ykzou1214--hy-worldplay"
func NewModalReplayClient(baseURL string, storage StorageClient) *ModalReplayClient {
	return &ModalReplayClient{
		baseURL: baseURL,
		storage: storage,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // Video generation takes longer
		},
	}
}

// modalReplayRequest is the request format for the Modal WorldPlay API
type modalReplayRequest struct {
	Prompt            string  `json:"prompt"`
	ImageBase64       *string `json:"image_base64,omitempty"`
	Resolution        string  `json:"resolution"` // "480p" or "720p"
	VideoLength       int     `json:"video_length,omitempty"`
	Pose              string  `json:"pose,omitempty"`
	ModelType         string  `json:"model_type,omitempty"` // "ar" or "bi"
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
}

// modalReplayResponse is the response format from the Modal WorldPlay API
type modalReplayResponse struct {
	VideoBase64 string `json:"video_base64"`
	Prompt      string `json:"prompt"`
	NumFrames   int    `json:"num_frames"`
	Pose        string `json:"pose"`
	Error       string `json:"error,omitempty"`
}

// GenerateReplay creates a video animation of a trajectory through the scene
func (c *ModalReplayClient) GenerateReplay(ctx context.Context, input models.ReplayInput) (*models.ReplayOutput, error) {
	startTime := time.Now()

	// 1. Build prompt from trajectory
	prompt := c.buildPromptFromTrajectory(input)

	// 2. Build camera pose string from trajectory
	poseStr := c.buildPoseString(input)

	// 3. Determine resolution
	resolution := input.Resolution
	if resolution == "" || (resolution != "480p" && resolution != "720p") {
		resolution = "480p" // Default to 480p
	}

	// 4. Fetch reference image if provided
	var imageBase64 *string
	if input.ReferenceImageKey != "" && c.storage != nil {
		data, _, err := c.storage.Download(ctx, "assets", input.ReferenceImageKey)
		if err == nil {
			encoded := base64.StdEncoding.EncodeToString(data)
			imageBase64 = &encoded
		}
	}

	// 5. Build request
	reqBody := modalReplayRequest{
		Prompt:            prompt,
		ImageBase64:       imageBase64,
		Resolution:        resolution,
		VideoLength:       input.FrameCount,
		Pose:              poseStr,
		ModelType:         "ar", // Auto-regressive mode
		NumInferenceSteps: 25,
	}

	if input.FrameCount == 0 {
		reqBody.VideoLength = 125 // Default
	}

	// 6. Make API call
	jsonBody, _ := json.Marshal(reqBody)
	url := c.baseURL + "-generate-video-api.modal.run"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// 7. Parse response
	var modalResp modalReplayResponse
	if err := json.Unmarshal(body, &modalResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if modalResp.Error != "" {
		return nil, fmt.Errorf("video generation failed: %s", modalResp.Error)
	}

	// 8. Save video to storage
	videoAssetKey, err := c.saveVideo(ctx, modalResp.VideoBase64, input.CaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}

	frameCount := modalResp.NumFrames
	if frameCount == 0 {
		frameCount = input.FrameCount
	}
	if frameCount == 0 {
		frameCount = 125
	}

	return &models.ReplayOutput{
		VideoAssetKey:  videoAssetKey,
		ThumbnailKey:   "", // TODO: Extract thumbnail from video
		FrameCount:     frameCount,
		FPS:            24,
		DurationMs:     int64(frameCount * 1000 / 24),
		Resolution:     resolution,
		ModelUsed:      "hy-world-1.5",
		GenerationTime: time.Since(startTime).Milliseconds(),
	}, nil
}

func (c *ModalReplayClient) buildPromptFromTrajectory(input models.ReplayInput) string {
	// Build a descriptive prompt for the scene
	prompt := "A crime scene interior"

	if input.SceneDescription != "" {
		prompt = input.SceneDescription
	}

	// Add trajectory context
	if input.TrajectoryDescription != "" {
		prompt += ". " + input.TrajectoryDescription
	}

	return prompt
}

func (c *ModalReplayClient) buildPoseString(input models.ReplayInput) string {
	// Convert trajectory to camera pose commands
	// Format: "action-duration,action-duration,..."
	// Actions: w (forward), s (back), a (left), d (right), up, down, left, right

	if input.CameraPose != "" {
		return input.CameraPose
	}

	// Default: move forward through the scene
	frameCount := input.FrameCount
	if frameCount == 0 {
		frameCount = 125
	}

	// Simple forward movement for now
	return fmt.Sprintf("w-%d", frameCount/4)
}

func (c *ModalReplayClient) saveVideo(ctx context.Context, videoBase64 string, caseID string) (string, error) {
	if c.storage == nil {
		// Return a placeholder key if no storage
		return fmt.Sprintf("cases/%s/replay/%s.mp4", caseID, uuid.New().String()), nil
	}

	// Decode base64 video
	videoData, err := base64.StdEncoding.DecodeString(videoBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode video: %w", err)
	}

	// Generate storage key
	assetKey := fmt.Sprintf("cases/%s/replay/%s.mp4", caseID, uuid.New().String())

	// Upload to storage
	if err := c.storage.Upload(ctx, "assets", assetKey, videoData, "video/mp4"); err != nil {
		return "", fmt.Errorf("failed to upload video: %w", err)
	}

	return assetKey, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func parsePose(data map[string]interface{}) models.Pose {
	pose := models.NewDefaultPose()

	if pos, ok := data["position"].([]interface{}); ok && len(pos) == 3 {
		pose.Position = [3]float64{
			toFloat64(pos[0]),
			toFloat64(pos[1]),
			toFloat64(pos[2]),
		}
	}

	if rot, ok := data["rotation"].([]interface{}); ok && len(rot) == 4 {
		pose.Rotation = [4]float64{
			toFloat64(rot[0]),
			toFloat64(rot[1]),
			toFloat64(rot[2]),
			toFloat64(rot[3]),
		}
	}

	return pose
}

func parseBBox(data map[string]interface{}) models.BoundingBox {
	bbox := models.BoundingBox{}

	if min, ok := data["min"].([]interface{}); ok && len(min) == 3 {
		bbox.Min = [3]float64{
			toFloat64(min[0]),
			toFloat64(min[1]),
			toFloat64(min[2]),
		}
	}

	if max, ok := data["max"].([]interface{}); ok && len(max) == 3 {
		bbox.Max = [3]float64{
			toFloat64(max[0]),
			toFloat64(max[1]),
			toFloat64(max[2]),
		}
	}

	return bbox
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}
