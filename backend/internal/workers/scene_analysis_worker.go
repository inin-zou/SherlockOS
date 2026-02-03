package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sherlockos/backend/internal/clients"
	"github.com/sherlockos/backend/internal/db"
	"github.com/sherlockos/backend/internal/models"
	"github.com/sherlockos/backend/internal/queue"
)

// SceneAnalysisWorker handles scene analysis jobs using Gemini 3 Pro Vision
type SceneAnalysisWorker struct {
	*BaseWorker
	client clients.SceneAnalysisClient
}

// NewSceneAnalysisWorker creates a new scene analysis worker
func NewSceneAnalysisWorker(database *db.DB, q queue.JobQueue, client clients.SceneAnalysisClient) *SceneAnalysisWorker {
	return &SceneAnalysisWorker{
		BaseWorker: NewBaseWorker(database, q),
		client:     client,
	}
}

// Type returns the job type this worker handles
func (w *SceneAnalysisWorker) Type() models.JobType {
	return models.JobTypeSceneAnalysis
}

// Process handles a scene analysis job
func (w *SceneAnalysisWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.SceneAnalysisInput
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Validate and set defaults
	if err := input.Validate(); err != nil {
		return NewFatalError(fmt.Errorf("invalid input: %w", err))
	}
	input.SetDefaults()

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Call scene analysis service (Gemini 3 Pro Vision)
	output, err := w.client.AnalyzeScene(ctx, input)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("scene analysis failed: %w", err))
	}

	// Update progress: analysis complete
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create commit with scene analysis results
	caseID, _ := uuid.Parse(input.CaseID)
	if err := w.createSceneAnalysisCommit(ctx, caseID, job.JobID, output); err != nil {
		fmt.Printf("Warning: failed to create scene analysis commit: %v\n", err)
	}

	// Update scene snapshot with detected objects
	if err := w.updateSceneSnapshot(ctx, caseID, output); err != nil {
		fmt.Printf("Warning: failed to update scene snapshot: %v\n", err)
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

// createSceneAnalysisCommit creates a commit for scene analysis results
func (w *SceneAnalysisWorker) createSceneAnalysisCommit(ctx context.Context, caseID, jobID uuid.UUID, output *models.SceneAnalysisOutput) error {
	if w.repo == nil {
		return nil
	}

	payload := map[string]interface{}{
		"job_id":             jobID.String(),
		"detected_objects":   output.DetectedObjects,
		"potential_evidence": output.PotentialEvidence,
		"scene_description":  output.SceneDescription,
		"anomalies":          output.Anomalies,
		"model_used":         output.ModelUsed,
		"analysis_time_ms":   output.AnalysisTime,
	}

	// Count suspicious objects
	suspiciousCount := 0
	for _, obj := range output.DetectedObjects {
		if obj.IsSuspicious {
			suspiciousCount++
		}
	}

	summary := fmt.Sprintf("Scene analysis: detected %d objects", len(output.DetectedObjects))
	if suspiciousCount > 0 {
		summary += fmt.Sprintf(" (%d suspicious)", suspiciousCount)
	}
	if len(output.PotentialEvidence) > 0 {
		summary += fmt.Sprintf(", %d potential evidence items", len(output.PotentialEvidence))
	}

	// Use reconstruction_update commit type since this updates the scene understanding
	commit, err := models.NewCommit(caseID, models.CommitTypeReconstructionUpdate, summary, payload)
	if err != nil {
		return err
	}

	// Get latest commit as parent
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	if latestCommit != nil {
		commit.SetParent(latestCommit.ID)
	}

	return w.repo.CreateCommit(ctx, commit)
}

// updateSceneSnapshot converts detected objects to SceneGraph and updates the snapshot
func (w *SceneAnalysisWorker) updateSceneSnapshot(ctx context.Context, caseID uuid.UUID, output *models.SceneAnalysisOutput) error {
	if w.repo == nil {
		return nil
	}

	// Get existing snapshot or create new one
	existing, _ := w.repo.GetSceneSnapshot(ctx, caseID)
	var sg *models.SceneGraph
	if existing != nil && existing.Scenegraph != nil {
		sg = existing.Scenegraph
	} else {
		sg = models.NewEmptySceneGraph()
	}

	// Convert DetectedObjects to SceneObjects
	for _, detected := range output.DetectedObjects {
		sceneObj := models.SceneObject{
			ID:         detected.ID,
			Type:       models.ObjectType(detected.Type),
			Label:      detected.Label,
			State:      "detected",
			Confidence: detected.Confidence,
			Pose: models.Pose{
				Position: [3]float64{0, 0, 0}, // Will be set by reconstruction
				Rotation: [4]float64{0, 0, 0, 1},
			},
			BBox: models.BoundingBox{
				Min: [3]float64{0, 0, 0},
				Max: [3]float64{1, 1, 1},
			},
			Metadata: map[string]interface{}{
				"notes":                detected.Notes,
				"is_suspicious":        detected.IsSuspicious,
				"position_description": detected.PositionDescription,
				"source_image_key":     detected.SourceImageKey,
			},
		}

		// Check if object already exists, update if so
		found := false
		for i, existing := range sg.Objects {
			if existing.ID == sceneObj.ID {
				sg.Objects[i] = sceneObj
				found = true
				break
			}
		}
		if !found {
			sg.Objects = append(sg.Objects, sceneObj)
		}
	}

	// Convert potential evidence to EvidenceCards
	for i, evidence := range output.PotentialEvidence {
		card := models.EvidenceCard{
			ID:          fmt.Sprintf("evidence_%d", i+1),
			Title:       evidence,
			Description: fmt.Sprintf("Potential evidence: %s", evidence),
			Confidence:  0.8,
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		}

		// Check if evidence already exists
		found := false
		for j, existing := range sg.Evidence {
			if existing.Title == card.Title {
				sg.Evidence[j] = card
				found = true
				break
			}
		}
		if !found {
			sg.Evidence = append(sg.Evidence, card)
		}
	}

	// Compute initial bounds from objects (will be refined by reconstruction)
	sg.Bounds = computeInitialBounds(sg.Objects)

	// Get latest commit for this case to use as commit_id
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	var commitID uuid.UUID
	if latestCommit != nil {
		commitID = latestCommit.ID
	}

	// Update snapshot
	snapshot := &models.SceneSnapshot{
		CaseID:     caseID,
		CommitID:   commitID,
		Scenegraph: sg,
		UpdatedAt:  time.Now().UTC(),
	}

	return w.repo.UpsertSceneSnapshot(ctx, snapshot)
}

// computeInitialBounds estimates scene bounds from detected objects
// Uses position descriptions to estimate room layout
func computeInitialBounds(objects []models.SceneObject) models.BoundingBox {
	if len(objects) == 0 {
		// Default bounds
		return models.BoundingBox{
			Min: [3]float64{-7, 0, -6},
			Max: [3]float64{7, 4, 6},
		}
	}

	// Analyze position descriptions to estimate room size
	hasWindow := false
	hasDoor := false
	hasFurniture := false
	evidenceCount := 0

	for _, obj := range objects {
		switch obj.Type {
		case models.ObjectTypeWindow:
			hasWindow = true
		case models.ObjectTypeDoor:
			hasDoor = true
		case models.ObjectTypeFurniture:
			hasFurniture = true
		case models.ObjectTypeEvidenceItem, models.ObjectTypeWeapon, models.ObjectTypeFootprint, models.ObjectTypeBloodstain:
			evidenceCount++
		}
	}

	// Base room size - typical office
	width := 7.0  // X dimension (left/right)
	depth := 6.0  // Z dimension (front/back)
	height := 4.0 // Y dimension (floor to ceiling)

	// Expand based on content
	if hasWindow && hasFurniture {
		width = 8.0
		depth = 7.0
	}
	if hasDoor {
		depth = 7.5 // Room with door entry tends to be deeper
	}
	if len(objects) > 10 {
		width = 9.0
		depth = 8.0
	}
	if evidenceCount > 5 {
		// Crime scene with lots of evidence - likely larger space
		width = 10.0
		depth = 9.0
	}

	return models.BoundingBox{
		Min: [3]float64{-width, 0, -depth},
		Max: [3]float64{width, height, depth},
	}
}
