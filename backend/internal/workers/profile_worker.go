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

// ProfileWorker handles suspect profile extraction jobs
type ProfileWorker struct {
	*BaseWorker
	client       clients.ProfileClient
	imageGenQueue queue.JobQueue // Optional: for triggering image gen jobs
}

// NewProfileWorker creates a new profile worker
func NewProfileWorker(database *db.DB, q queue.JobQueue, client clients.ProfileClient) *ProfileWorker {
	return &ProfileWorker{
		BaseWorker:    NewBaseWorker(database, q),
		client:        client,
		imageGenQueue: q,
	}
}

// Type returns the job type this worker handles
func (w *ProfileWorker) Type() models.JobType {
	return models.JobTypeProfile
}

// Process handles a profile extraction job
func (w *ProfileWorker) Process(ctx context.Context, job *queue.JobMessage) error {
	// Parse input
	var input models.ProfileInput
	if err := json.Unmarshal(job.Input, &input); err != nil {
		return NewFatalError(fmt.Errorf("failed to parse input: %w", err))
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return NewFatalError(fmt.Errorf("invalid input: %w", err))
	}

	// Update progress: starting
	w.UpdateJobProgress(ctx, job.JobID, 10)

	// Get existing profile if available
	caseID, _ := uuid.Parse(input.CaseID)
	var existingAttrs *models.SuspectAttributes
	if w.repo != nil {
		existingProfile, _ := w.repo.GetSuspectProfile(ctx, caseID)
		if existingProfile != nil {
			existingAttrs = existingProfile.Attributes
		}
	}

	// Update progress: calling AI
	w.UpdateJobProgress(ctx, job.JobID, 30)

	// Extract profile using AI
	newAttrs, err := w.client.ExtractProfile(ctx, input.Statements, existingAttrs)
	if err != nil {
		w.MarkJobFailed(ctx, job.JobID, err)
		return NewRetryableError(fmt.Errorf("profile extraction failed: %w", err))
	}

	// Update progress: processing complete
	w.UpdateJobProgress(ctx, job.JobID, 60)

	// Merge with existing attributes
	mergedAttrs := w.mergeAttributes(existingAttrs, newAttrs, input.Statements)

	// Detect conflicts
	conflicts := w.detectConflicts(input.Statements, newAttrs)

	// Update progress: saving
	w.UpdateJobProgress(ctx, job.JobID, 80)

	// Create profile update commit
	commitID, err := w.createProfileCommit(ctx, caseID, job.JobID, mergedAttrs, conflicts)
	if err != nil {
		fmt.Printf("Warning: failed to create profile commit: %v\n", err)
	}

	// Update suspect profile in database
	if err := w.updateSuspectProfile(ctx, caseID, commitID, mergedAttrs); err != nil {
		fmt.Printf("Warning: failed to update suspect profile: %v\n", err)
	}

	// Check if we should trigger portrait generation
	imageGenTriggered := false
	var imageGenJobID string
	if w.shouldTriggerImageGen(mergedAttrs) {
		jobID, err := w.triggerImageGenJob(ctx, caseID, mergedAttrs)
		if err == nil {
			imageGenTriggered = true
			imageGenJobID = jobID
		}
	}

	// Build output
	output := &models.ProfileOutput{
		Attributes:        mergedAttrs,
		ExtractedFacts:    w.extractFacts(input.Statements, newAttrs),
		Conflicts:         conflicts,
		ImageGenTriggered: imageGenTriggered,
		ImageGenJobID:     imageGenJobID,
	}

	// Mark job as done
	w.UpdateJobProgress(ctx, job.JobID, 100)
	w.MarkJobDone(ctx, job.JobID, output)

	return nil
}

// mergeAttributes merges new attributes with existing ones
func (w *ProfileWorker) mergeAttributes(existing, new *models.SuspectAttributes, statements []models.WitnessStatementInput) *models.SuspectAttributes {
	if existing == nil {
		return new
	}
	if new == nil {
		return existing
	}

	result := models.NewEmptySuspectAttributes()

	// Merge each attribute, preferring higher confidence
	result.AgeRange = mergeRangeAttr(existing.AgeRange, new.AgeRange)
	result.HeightRangeCm = mergeRangeAttr(existing.HeightRangeCm, new.HeightRangeCm)
	result.Build = mergeStringAttr(existing.Build, new.Build)
	result.SkinTone = mergeStringAttr(existing.SkinTone, new.SkinTone)
	result.FacialHair = mergeStringAttr(existing.FacialHair, new.FacialHair)
	result.Glasses = mergeStringAttr(existing.Glasses, new.Glasses)
	result.Hair = mergeHairAttr(existing.Hair, new.Hair)

	// Merge distinctive features (combine lists)
	featureSet := make(map[string]models.FeatureAttribute)
	for _, f := range existing.DistinctiveFeatures {
		featureSet[f.Description] = f
	}
	for _, f := range new.DistinctiveFeatures {
		if existingF, ok := featureSet[f.Description]; ok {
			// Average confidence if same feature
			featureSet[f.Description] = models.FeatureAttribute{
				Description: f.Description,
				Confidence:  (existingF.Confidence + f.Confidence) / 2,
			}
		} else {
			featureSet[f.Description] = f
		}
	}
	for _, f := range featureSet {
		result.DistinctiveFeatures = append(result.DistinctiveFeatures, f)
	}

	return result
}

func mergeHairAttr(a, b *models.HairAttribute) *models.HairAttribute {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if b.Confidence > a.Confidence {
		return b
	}
	return a
}

func mergeRangeAttr(a, b *models.RangeAttribute) *models.RangeAttribute {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	// Return higher confidence
	if b.Confidence > a.Confidence {
		return b
	}
	return a
}

func mergeStringAttr(a, b *models.StringAttribute) *models.StringAttribute {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if b.Confidence > a.Confidence {
		return b
	}
	return a
}

// detectConflicts finds conflicting attribute values from different statements
func (w *ProfileWorker) detectConflicts(statements []models.WitnessStatementInput, attrs *models.SuspectAttributes) []models.AttributeConflict {
	// In a full implementation, this would compare values from different statements
	// For now, return empty conflicts
	return []models.AttributeConflict{}
}

// extractFacts extracts individual facts from statements
func (w *ProfileWorker) extractFacts(statements []models.WitnessStatementInput, attrs *models.SuspectAttributes) []models.ExtractedFact {
	var facts []models.ExtractedFact

	if attrs == nil {
		return facts
	}

	if attrs.AgeRange != nil {
		facts = append(facts, models.ExtractedFact{
			Attribute:  "age_range",
			Value:      fmt.Sprintf("%d-%d", int(attrs.AgeRange.Min), int(attrs.AgeRange.Max)),
			Confidence: attrs.AgeRange.Confidence,
		})
	}
	if attrs.Build != nil {
		facts = append(facts, models.ExtractedFact{
			Attribute:  "build",
			Value:      attrs.Build.Value,
			Confidence: attrs.Build.Confidence,
		})
	}
	if attrs.Hair != nil && attrs.Hair.Color != "" {
		facts = append(facts, models.ExtractedFact{
			Attribute:  "hair_color",
			Value:      attrs.Hair.Color,
			Confidence: attrs.Hair.Confidence,
		})
	}
	if attrs.SkinTone != nil {
		facts = append(facts, models.ExtractedFact{
			Attribute:  "skin_tone",
			Value:      attrs.SkinTone.Value,
			Confidence: attrs.SkinTone.Confidence,
		})
	}

	return facts
}

// shouldTriggerImageGen determines if we have enough attributes for portrait generation
func (w *ProfileWorker) shouldTriggerImageGen(attrs *models.SuspectAttributes) bool {
	if attrs == nil {
		return false
	}

	// Require at least a few attributes with decent confidence
	attrCount := 0
	if attrs.AgeRange != nil && attrs.AgeRange.Confidence > 0.5 {
		attrCount++
	}
	if attrs.Build != nil && attrs.Build.Confidence > 0.5 {
		attrCount++
	}
	if attrs.Hair != nil && attrs.Hair.Confidence > 0.5 {
		attrCount++
	}
	if attrs.SkinTone != nil && attrs.SkinTone.Confidence > 0.5 {
		attrCount++
	}

	return attrCount >= 3
}

// triggerImageGenJob creates an image generation job for portrait
func (w *ProfileWorker) triggerImageGenJob(ctx context.Context, caseID uuid.UUID, attrs *models.SuspectAttributes) (string, error) {
	if w.repo == nil || w.imageGenQueue == nil {
		return "", nil
	}

	input := models.ImageGenInput{
		CaseID:        caseID.String(),
		GenType:       models.ImageGenTypePortrait,
		PortraitAttrs: attrs,
		Resolution:    "1k",
	}

	job, err := models.NewJob(caseID, models.JobTypeImageGen, input)
	if err != nil {
		return "", err
	}

	if err := w.repo.CreateJob(ctx, job); err != nil {
		return "", err
	}

	if err := w.imageGenQueue.Enqueue(ctx, job); err != nil {
		return "", err
	}

	return job.ID.String(), nil
}

// createProfileCommit creates a commit for profile update
func (w *ProfileWorker) createProfileCommit(ctx context.Context, caseID, jobID uuid.UUID, attrs *models.SuspectAttributes, conflicts []models.AttributeConflict) (uuid.UUID, error) {
	if w.repo == nil {
		return uuid.Nil, nil
	}

	payload := map[string]interface{}{
		"job_id":     jobID.String(),
		"attributes": attrs,
		"conflicts":  conflicts,
	}

	summary := "Updated suspect profile from witness statements"

	commit, err := models.NewCommit(caseID, models.CommitTypeProfileUpdate, summary, payload)
	if err != nil {
		return uuid.Nil, err
	}

	// Get latest commit as parent
	latestCommit, _ := w.repo.GetLatestCommit(ctx, caseID)
	if latestCommit != nil {
		commit.SetParent(latestCommit.ID)
	}

	if err := w.repo.CreateCommit(ctx, commit); err != nil {
		return uuid.Nil, err
	}

	return commit.ID, nil
}

// updateSuspectProfile updates the suspect profile in database
func (w *ProfileWorker) updateSuspectProfile(ctx context.Context, caseID, commitID uuid.UUID, attrs *models.SuspectAttributes) error {
	if w.repo == nil {
		return nil
	}

	profile := &models.SuspectProfile{
		CaseID:     caseID,
		CommitID:   commitID,
		Attributes: attrs,
		UpdatedAt:  time.Now().UTC(),
	}

	return w.repo.UpsertSuspectProfile(ctx, profile)
}
