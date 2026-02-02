package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/sherlockos/backend/internal/models"
)

// Repository provides database operations
type Repository struct {
	db *DB
}

// NewRepository creates a new repository
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// ============================================
// CASES
// ============================================

// CreateCase creates a new case
func (r *Repository) CreateCase(ctx context.Context, c *models.Case) error {
	query := `
		INSERT INTO cases (id, title, description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query, c.ID, c.Title, c.Description, c.CreatedBy, c.CreatedAt)
	return err
}

// GetCase retrieves a case by ID
func (r *Repository) GetCase(ctx context.Context, id uuid.UUID) (*models.Case, error) {
	query := `
		SELECT id, title, description, created_by, created_at
		FROM cases WHERE id = $1
	`
	var c models.Case
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListCases returns all cases with pagination
func (r *Repository) ListCases(ctx context.Context, limit int, cursor *time.Time) ([]*models.Case, error) {
	var query string
	var args []interface{}

	if cursor != nil {
		query = `
			SELECT id, title, description, created_by, created_at
			FROM cases WHERE created_at < $1
			ORDER BY created_at DESC LIMIT $2
		`
		args = []interface{}{cursor, limit}
	} else {
		query = `
			SELECT id, title, description, created_by, created_at
			FROM cases ORDER BY created_at DESC LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []*models.Case
	for rows.Next() {
		var c models.Case
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		cases = append(cases, &c)
	}
	return cases, nil
}

// ============================================
// COMMITS
// ============================================

// CreateCommit creates a new commit
func (r *Repository) CreateCommit(ctx context.Context, c *models.Commit) error {
	query := `
		INSERT INTO commits (id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.CaseID, c.ParentCommitID, c.BranchID, c.Type, c.Summary, c.Payload, c.CreatedBy, c.CreatedAt,
	)
	return err
}

// GetCommit retrieves a commit by ID
func (r *Repository) GetCommit(ctx context.Context, id uuid.UUID) (*models.Commit, error) {
	query := `
		SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at
		FROM commits WHERE id = $1
	`
	var c models.Commit
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.CaseID, &c.ParentCommitID, &c.BranchID, &c.Type, &c.Summary, &c.Payload, &c.CreatedBy, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCommitsByCase returns commits for a case with pagination
func (r *Repository) GetCommitsByCase(ctx context.Context, caseID uuid.UUID, limit int, cursor *time.Time) ([]*models.Commit, error) {
	var query string
	var args []interface{}

	if cursor != nil {
		query = `
			SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at
			FROM commits WHERE case_id = $1 AND created_at < $2
			ORDER BY created_at DESC LIMIT $3
		`
		args = []interface{}{caseID, cursor, limit}
	} else {
		query = `
			SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at
			FROM commits WHERE case_id = $1
			ORDER BY created_at DESC LIMIT $2
		`
		args = []interface{}{caseID, limit}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*models.Commit
	for rows.Next() {
		var c models.Commit
		if err := rows.Scan(&c.ID, &c.CaseID, &c.ParentCommitID, &c.BranchID, &c.Type, &c.Summary, &c.Payload, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		commits = append(commits, &c)
	}
	return commits, nil
}

// GetLatestCommit returns the most recent commit for a case
func (r *Repository) GetLatestCommit(ctx context.Context, caseID uuid.UUID) (*models.Commit, error) {
	query := `
		SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at
		FROM commits WHERE case_id = $1
		ORDER BY created_at DESC LIMIT 1
	`
	var c models.Commit
	err := r.db.Pool.QueryRow(ctx, query, caseID).Scan(
		&c.ID, &c.CaseID, &c.ParentCommitID, &c.BranchID, &c.Type, &c.Summary, &c.Payload, &c.CreatedBy, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ============================================
// BRANCHES
// ============================================

// CreateBranch creates a new branch
func (r *Repository) CreateBranch(ctx context.Context, b *models.Branch) error {
	query := `
		INSERT INTO branches (id, case_id, name, base_commit_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query, b.ID, b.CaseID, b.Name, b.BaseCommitID, b.CreatedAt)
	return err
}

// GetBranch retrieves a branch by ID
func (r *Repository) GetBranch(ctx context.Context, id uuid.UUID) (*models.Branch, error) {
	query := `
		SELECT id, case_id, name, base_commit_id, created_at
		FROM branches WHERE id = $1
	`
	var b models.Branch
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.CaseID, &b.Name, &b.BaseCommitID, &b.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBranchesByCase returns all branches for a case
func (r *Repository) GetBranchesByCase(ctx context.Context, caseID uuid.UUID) ([]*models.Branch, error) {
	query := `
		SELECT id, case_id, name, base_commit_id, created_at
		FROM branches WHERE case_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(&b.ID, &b.CaseID, &b.Name, &b.BaseCommitID, &b.CreatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, &b)
	}
	return branches, nil
}

// ============================================
// JOBS
// ============================================

// CreateJob creates a new job
func (r *Repository) CreateJob(ctx context.Context, j *models.Job) error {
	query := `
		INSERT INTO jobs (id, case_id, type, status, progress, input, output, error, idempotency_key, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	// Use nil for empty idempotency key to allow multiple jobs without keys
	var idempotencyKey interface{}
	if j.IdempotencyKey != "" {
		idempotencyKey = j.IdempotencyKey
	}
	_, err := r.db.Pool.Exec(ctx, query,
		j.ID, j.CaseID, j.Type, j.Status, j.Progress, j.Input, j.Output, j.Error, idempotencyKey, j.RetryCount, j.CreatedAt, j.UpdatedAt,
	)
	return err
}

// GetJob retrieves a job by ID
func (r *Repository) GetJob(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	query := `
		SELECT id, case_id, type, status, progress, input, output, error, idempotency_key, retry_count, created_at, updated_at
		FROM jobs WHERE id = $1
	`
	var j models.Job
	var idempotencyKey *string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&j.ID, &j.CaseID, &j.Type, &j.Status, &j.Progress, &j.Input, &j.Output, &j.Error, &idempotencyKey, &j.RetryCount, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if idempotencyKey != nil {
		j.IdempotencyKey = *idempotencyKey
	}
	return &j, nil
}

// GetJobByIdempotencyKey retrieves a job by idempotency key
func (r *Repository) GetJobByIdempotencyKey(ctx context.Context, key string) (*models.Job, error) {
	query := `
		SELECT id, case_id, type, status, progress, input, output, error, idempotency_key, retry_count, created_at, updated_at
		FROM jobs WHERE idempotency_key = $1
	`
	var j models.Job
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(
		&j.ID, &j.CaseID, &j.Type, &j.Status, &j.Progress, &j.Input, &j.Output, &j.Error, &j.IdempotencyKey, &j.RetryCount, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// UpdateJobStatus updates job status and progress
func (r *Repository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, progress int) error {
	query := `UPDATE jobs SET status = $2, progress = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, status, progress)
	return err
}

// UpdateJobOutput updates job output when complete
func (r *Repository) UpdateJobOutput(ctx context.Context, id uuid.UUID, output interface{}) error {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return err
	}
	query := `UPDATE jobs SET status = 'done', progress = 100, output = $2, updated_at = NOW() WHERE id = $1`
	_, err = r.db.Pool.Exec(ctx, query, id, outputJSON)
	return err
}

// UpdateJobError marks job as failed with error message
func (r *Repository) UpdateJobError(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE jobs SET status = 'failed', error = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, errMsg)
	return err
}

// UpdateJobHeartbeat updates the updated_at timestamp
func (r *Repository) UpdateJobHeartbeat(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE jobs SET updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// GetQueuedJobs returns jobs in queued status for a specific type
func (r *Repository) GetQueuedJobs(ctx context.Context, jobType models.JobType, limit int) ([]*models.Job, error) {
	query := `
		SELECT id, case_id, type, status, progress, input, output, error, COALESCE(idempotency_key, ''), retry_count, created_at, updated_at
		FROM jobs WHERE type = $1 AND status = 'queued'
		ORDER BY created_at ASC LIMIT $2
	`
	rows, err := r.db.Pool.Query(ctx, query, jobType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.CaseID, &j.Type, &j.Status, &j.Progress, &j.Input, &j.Output, &j.Error, &j.IdempotencyKey, &j.RetryCount, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, nil
}

// GetZombieJobs returns jobs that are running but haven't been updated recently
func (r *Repository) GetZombieJobs(ctx context.Context, timeout time.Duration) ([]*models.Job, error) {
	query := `
		SELECT id, case_id, type, status, progress, input, output, error, COALESCE(idempotency_key, ''), retry_count, created_at, updated_at
		FROM jobs WHERE status = 'running' AND updated_at < NOW() - $1::interval
	`
	rows, err := r.db.Pool.Query(ctx, query, fmt.Sprintf("%d seconds", int(timeout.Seconds())))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.CaseID, &j.Type, &j.Status, &j.Progress, &j.Input, &j.Output, &j.Error, &j.IdempotencyKey, &j.RetryCount, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, &j)
	}
	return jobs, nil
}

// IncrementJobRetry increments retry count and requeues job
func (r *Repository) IncrementJobRetry(ctx context.Context, id uuid.UUID, maxRetries int) (bool, error) {
	query := `
		UPDATE jobs SET
			retry_count = retry_count + 1,
			status = CASE WHEN retry_count + 1 >= $2 THEN 'failed' ELSE 'queued' END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING status
	`
	var status models.JobStatus
	err := r.db.Pool.QueryRow(ctx, query, id, maxRetries).Scan(&status)
	if err != nil {
		return false, err
	}
	return status == models.JobStatusQueued, nil
}

// ============================================
// SCENE SNAPSHOTS
// ============================================

// UpsertSceneSnapshot creates or updates scene snapshot
func (r *Repository) UpsertSceneSnapshot(ctx context.Context, ss *models.SceneSnapshot) error {
	sgJSON, err := json.Marshal(ss.Scenegraph)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO scene_snapshots (case_id, commit_id, scenegraph, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (case_id) DO UPDATE SET
			commit_id = EXCLUDED.commit_id,
			scenegraph = EXCLUDED.scenegraph,
			updated_at = EXCLUDED.updated_at
	`
	_, err = r.db.Pool.Exec(ctx, query, ss.CaseID, ss.CommitID, sgJSON, ss.UpdatedAt)
	return err
}

// GetSceneSnapshot retrieves scene snapshot for a case
func (r *Repository) GetSceneSnapshot(ctx context.Context, caseID uuid.UUID) (*models.SceneSnapshot, error) {
	query := `
		SELECT case_id, commit_id, scenegraph, updated_at
		FROM scene_snapshots WHERE case_id = $1
	`
	var ss models.SceneSnapshot
	var sgJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, caseID).Scan(&ss.CaseID, &ss.CommitID, &sgJSON, &ss.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ss.Scenegraph = &models.SceneGraph{}
	if err := json.Unmarshal(sgJSON, ss.Scenegraph); err != nil {
		return nil, err
	}
	return &ss, nil
}

// ============================================
// SUSPECT PROFILES
// ============================================

// UpsertSuspectProfile creates or updates suspect profile
func (r *Repository) UpsertSuspectProfile(ctx context.Context, sp *models.SuspectProfile) error {
	attrsJSON, err := json.Marshal(sp.Attributes)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO suspect_profiles (case_id, commit_id, attributes, portrait_asset_key, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (case_id) DO UPDATE SET
			commit_id = EXCLUDED.commit_id,
			attributes = EXCLUDED.attributes,
			portrait_asset_key = EXCLUDED.portrait_asset_key,
			updated_at = EXCLUDED.updated_at
	`
	_, err = r.db.Pool.Exec(ctx, query, sp.CaseID, sp.CommitID, attrsJSON, sp.PortraitAssetKey, sp.UpdatedAt)
	return err
}

// GetSuspectProfile retrieves suspect profile for a case
func (r *Repository) GetSuspectProfile(ctx context.Context, caseID uuid.UUID) (*models.SuspectProfile, error) {
	query := `
		SELECT case_id, commit_id, attributes, portrait_asset_key, updated_at
		FROM suspect_profiles WHERE case_id = $1
	`
	var sp models.SuspectProfile
	var attrsJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, caseID).Scan(&sp.CaseID, &sp.CommitID, &attrsJSON, &sp.PortraitAssetKey, &sp.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sp.Attributes = &models.SuspectAttributes{}
	if err := json.Unmarshal(attrsJSON, sp.Attributes); err != nil {
		return nil, err
	}
	return &sp, nil
}

// ============================================
// ASSETS
// ============================================

// CreateAsset creates a new asset
func (r *Repository) CreateAsset(ctx context.Context, a *models.Asset) error {
	metaJSON, err := json.Marshal(a.Metadata)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO assets (id, case_id, kind, storage_key, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.Pool.Exec(ctx, query, a.ID, a.CaseID, a.Kind, a.StorageKey, metaJSON, a.CreatedAt)
	return err
}

// GetAsset retrieves an asset by ID
func (r *Repository) GetAsset(ctx context.Context, id uuid.UUID) (*models.Asset, error) {
	query := `
		SELECT id, case_id, kind, storage_key, metadata, created_at
		FROM assets WHERE id = $1
	`
	var a models.Asset
	var metaJSON []byte
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&a.ID, &a.CaseID, &a.Kind, &a.StorageKey, &metaJSON, &a.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(metaJSON, &a.Metadata); err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAssetsByCase returns all assets for a case
func (r *Repository) GetAssetsByCase(ctx context.Context, caseID uuid.UUID, kind *models.AssetKind) ([]*models.Asset, error) {
	var query string
	var args []interface{}

	if kind != nil {
		query = `
			SELECT id, case_id, kind, storage_key, metadata, created_at
			FROM assets WHERE case_id = $1 AND kind = $2
			ORDER BY created_at DESC
		`
		args = []interface{}{caseID, *kind}
	} else {
		query = `
			SELECT id, case_id, kind, storage_key, metadata, created_at
			FROM assets WHERE case_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{caseID}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*models.Asset
	for rows.Next() {
		var a models.Asset
		var metaJSON []byte
		if err := rows.Scan(&a.ID, &a.CaseID, &a.Kind, &a.StorageKey, &metaJSON, &a.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metaJSON, &a.Metadata); err != nil {
			return nil, err
		}
		assets = append(assets, &a)
	}
	return assets, nil
}

// ============================================
// COMMIT DIFF & REPLAY
// ============================================

// SceneGraphDiff represents the difference between two SceneGraphs
type SceneGraphDiff struct {
	ObjectsAdded    []models.SceneObject   `json:"objects_added"`
	ObjectsUpdated  []ObjectUpdate         `json:"objects_updated"`
	ObjectsRemoved  []string               `json:"objects_removed"` // IDs of removed objects
	EvidenceAdded   []models.EvidenceCard  `json:"evidence_added"`
	EvidenceUpdated []EvidenceUpdate       `json:"evidence_updated"`
	EvidenceRemoved []string               `json:"evidence_removed"` // IDs of removed evidence
}

// ObjectUpdate represents an update to an object
type ObjectUpdate struct {
	ID     string             `json:"id"`
	Before models.SceneObject `json:"before"`
	After  models.SceneObject `json:"after"`
}

// EvidenceUpdate represents an update to an evidence card
type EvidenceUpdate struct {
	ID     string              `json:"id"`
	Before models.EvidenceCard `json:"before"`
	After  models.EvidenceCard `json:"after"`
}

// GetCommitDiff computes the difference between two commits' SceneGraphs
func (r *Repository) GetCommitDiff(ctx context.Context, fromCommitID, toCommitID uuid.UUID) (*SceneGraphDiff, error) {
	// Get commits
	fromCommit, err := r.GetCommit(ctx, fromCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from commit: %w", err)
	}
	if fromCommit == nil {
		return nil, fmt.Errorf("from commit not found")
	}

	toCommit, err := r.GetCommit(ctx, toCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to commit: %w", err)
	}
	if toCommit == nil {
		return nil, fmt.Errorf("to commit not found")
	}

	// Replay to get SceneGraphs at each commit
	fromGraph, err := r.ReplayToCommit(ctx, fromCommit.CaseID, fromCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to replay to from commit: %w", err)
	}

	toGraph, err := r.ReplayToCommit(ctx, toCommit.CaseID, toCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to replay to to commit: %w", err)
	}

	// Compute diff
	return ComputeSceneGraphDiff(fromGraph, toGraph), nil
}

// ComputeSceneGraphDiff computes the difference between two SceneGraphs
func ComputeSceneGraphDiff(from, to *models.SceneGraph) *SceneGraphDiff {
	diff := &SceneGraphDiff{
		ObjectsAdded:    []models.SceneObject{},
		ObjectsUpdated:  []ObjectUpdate{},
		ObjectsRemoved:  []string{},
		EvidenceAdded:   []models.EvidenceCard{},
		EvidenceUpdated: []EvidenceUpdate{},
		EvidenceRemoved: []string{},
	}

	// Handle nil inputs
	if from == nil {
		from = models.NewEmptySceneGraph()
	}
	if to == nil {
		to = models.NewEmptySceneGraph()
	}

	// Build maps for O(1) lookup
	fromObjects := make(map[string]models.SceneObject)
	for _, obj := range from.Objects {
		fromObjects[obj.ID] = obj
	}

	toObjects := make(map[string]models.SceneObject)
	for _, obj := range to.Objects {
		toObjects[obj.ID] = obj
	}

	// Find added and updated objects
	for id, toObj := range toObjects {
		if fromObj, exists := fromObjects[id]; exists {
			// Check if updated (compare key fields)
			if !objectsEqual(fromObj, toObj) {
				diff.ObjectsUpdated = append(diff.ObjectsUpdated, ObjectUpdate{
					ID:     id,
					Before: fromObj,
					After:  toObj,
				})
			}
		} else {
			diff.ObjectsAdded = append(diff.ObjectsAdded, toObj)
		}
	}

	// Find removed objects
	for id := range fromObjects {
		if _, exists := toObjects[id]; !exists {
			diff.ObjectsRemoved = append(diff.ObjectsRemoved, id)
		}
	}

	// Build maps for evidence
	fromEvidence := make(map[string]models.EvidenceCard)
	for _, ev := range from.Evidence {
		fromEvidence[ev.ID] = ev
	}

	toEvidence := make(map[string]models.EvidenceCard)
	for _, ev := range to.Evidence {
		toEvidence[ev.ID] = ev
	}

	// Find added and updated evidence
	for id, toEv := range toEvidence {
		if fromEv, exists := fromEvidence[id]; exists {
			if !evidenceEqual(fromEv, toEv) {
				diff.EvidenceUpdated = append(diff.EvidenceUpdated, EvidenceUpdate{
					ID:     id,
					Before: fromEv,
					After:  toEv,
				})
			}
		} else {
			diff.EvidenceAdded = append(diff.EvidenceAdded, toEv)
		}
	}

	// Find removed evidence
	for id := range fromEvidence {
		if _, exists := toEvidence[id]; !exists {
			diff.EvidenceRemoved = append(diff.EvidenceRemoved, id)
		}
	}

	return diff
}

// objectsEqual compares two SceneObjects for equality
func objectsEqual(a, b models.SceneObject) bool {
	// Compare key fields that indicate a meaningful change
	if a.Label != b.Label ||
		a.Type != b.Type ||
		a.State != b.State ||
		a.Confidence != b.Confidence {
		return false
	}

	// Compare pose
	if a.Pose.Position != b.Pose.Position ||
		a.Pose.Rotation != b.Pose.Rotation {
		return false
	}

	return true
}

// evidenceEqual compares two EvidenceCards for equality
func evidenceEqual(a, b models.EvidenceCard) bool {
	return a.Title == b.Title &&
		a.Description == b.Description &&
		a.Confidence == b.Confidence
}

// ReplayToCommit reconstructs the SceneGraph at a specific commit by replaying all commits
func (r *Repository) ReplayToCommit(ctx context.Context, caseID, targetCommitID uuid.UUID) (*models.SceneGraph, error) {
	// Get the commit chain from the beginning to the target commit
	commits, err := r.getCommitChain(ctx, caseID, targetCommitID)
	if err != nil {
		return nil, err
	}

	// Start with empty SceneGraph
	sg := models.NewEmptySceneGraph()

	// Apply each commit in order
	for _, commit := range commits {
		if err := applyCommitToSceneGraph(sg, commit); err != nil {
			return nil, fmt.Errorf("failed to apply commit %s: %w", commit.ID, err)
		}
	}

	return sg, nil
}

// getCommitChain retrieves all commits from the beginning to the target commit
func (r *Repository) getCommitChain(ctx context.Context, caseID, targetCommitID uuid.UUID) ([]*models.Commit, error) {
	// Get all commits for the case ordered by creation time
	query := `
		WITH RECURSIVE commit_chain AS (
			-- Start from target commit
			SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at, 0 as depth
			FROM commits
			WHERE id = $1 AND case_id = $2

			UNION ALL

			-- Walk up the parent chain
			SELECT c.id, c.case_id, c.parent_commit_id, c.branch_id, c.type, c.summary, c.payload, c.created_by, c.created_at, cc.depth + 1
			FROM commits c
			INNER JOIN commit_chain cc ON c.id = cc.parent_commit_id
		)
		SELECT id, case_id, parent_commit_id, branch_id, type, summary, payload, created_by, created_at
		FROM commit_chain
		ORDER BY depth DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, targetCommitID, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*models.Commit
	for rows.Next() {
		var c models.Commit
		if err := rows.Scan(&c.ID, &c.CaseID, &c.ParentCommitID, &c.BranchID, &c.Type, &c.Summary, &c.Payload, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		commits = append(commits, &c)
	}

	return commits, nil
}

// applyCommitToSceneGraph applies a commit's changes to a SceneGraph
func applyCommitToSceneGraph(sg *models.SceneGraph, commit *models.Commit) error {
	// Parse commit payload
	var payload models.CommitPayload
	if err := json.Unmarshal(commit.Payload, &payload); err != nil {
		// Not all commits have standard payload structure, skip if can't parse
		return nil
	}

	// Handle reconstruction updates - look for scenegraph in payload
	if commit.Type == models.CommitTypeReconstructionUpdate {
		var reconPayload struct {
			SceneGraph *models.SceneGraph `json:"scenegraph"`
		}
		if err := json.Unmarshal(commit.Payload, &reconPayload); err == nil && reconPayload.SceneGraph != nil {
			// Merge objects from reconstruction
			existingObjects := make(map[string]int)
			for i, obj := range sg.Objects {
				existingObjects[obj.ID] = i
			}
			for _, newObj := range reconPayload.SceneGraph.Objects {
				if idx, exists := existingObjects[newObj.ID]; exists {
					sg.Objects[idx] = newObj
				} else {
					sg.Objects = append(sg.Objects, newObj)
				}
			}
			// Merge evidence
			existingEvidence := make(map[string]int)
			for i, ev := range sg.Evidence {
				existingEvidence[ev.ID] = i
			}
			for _, newEv := range reconPayload.SceneGraph.Evidence {
				if idx, exists := existingEvidence[newEv.ID]; exists {
					sg.Evidence[idx] = newEv
				} else {
					sg.Evidence = append(sg.Evidence, newEv)
				}
			}
			// Update bounds
			sg.Bounds = reconPayload.SceneGraph.Bounds
		}
	}

	// Handle manual edits
	if commit.Type == models.CommitTypeManualEdit && payload.Changes != nil {
		// Remove objects marked for removal
		if len(payload.Changes.ObjectsRemoved) > 0 {
			removeSet := make(map[string]bool)
			for _, id := range payload.Changes.ObjectsRemoved {
				removeSet[id] = true
			}
			var newObjects []models.SceneObject
			for _, obj := range sg.Objects {
				if !removeSet[obj.ID] {
					newObjects = append(newObjects, obj)
				}
			}
			sg.Objects = newObjects
		}
	}

	return nil
}
