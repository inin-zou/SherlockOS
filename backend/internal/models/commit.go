package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Commit represents a timeline entry (append-only)
type Commit struct {
	ID             uuid.UUID       `json:"id"`
	CaseID         uuid.UUID       `json:"case_id"`
	ParentCommitID *uuid.UUID      `json:"parent_commit_id,omitempty"`
	BranchID       *uuid.UUID      `json:"branch_id,omitempty"`
	Type           CommitType      `json:"type"`
	Summary        string          `json:"summary"`
	Payload        json.RawMessage `json:"payload"`
	CreatedBy      *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Validate checks if the Commit is valid
func (c *Commit) Validate() error {
	if c.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if !c.Type.IsValid() {
		return errors.New("invalid commit type")
	}
	if c.Summary == "" {
		return errors.New("summary is required")
	}
	if len(c.Summary) > 500 {
		return errors.New("summary must be 500 characters or less")
	}
	return nil
}

// NewCommit creates a new Commit with generated ID and timestamp
func NewCommit(caseID uuid.UUID, commitType CommitType, summary string, payload interface{}) (*Commit, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Commit{
		ID:        uuid.New(),
		CaseID:    caseID,
		Type:      commitType,
		Summary:   summary,
		Payload:   payloadBytes,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// SetParent sets the parent commit ID
func (c *Commit) SetParent(parentID uuid.UUID) {
	c.ParentCommitID = &parentID
}

// SetBranch sets the branch ID
func (c *Commit) SetBranch(branchID uuid.UUID) {
	c.BranchID = &branchID
}

// CommitPayload represents the common structure for commit payloads
type CommitPayload struct {
	JobID      string                 `json:"job_id,omitempty"`
	AssetKeys  []string               `json:"asset_keys,omitempty"`
	Changes    *CommitChanges         `json:"changes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CommitChanges tracks what changed in a commit
type CommitChanges struct {
	ObjectsAdded    []string `json:"objects_added,omitempty"`
	ObjectsUpdated  []string `json:"objects_updated,omitempty"`
	ObjectsRemoved  []string `json:"objects_removed,omitempty"`
	EvidenceAdded   []string `json:"evidence_added,omitempty"`
	EvidenceUpdated []string `json:"evidence_updated,omitempty"`
}

// Branch represents a hypothesis branch
type Branch struct {
	ID           uuid.UUID `json:"id"`
	CaseID       uuid.UUID `json:"case_id"`
	Name         string    `json:"name"`
	BaseCommitID uuid.UUID `json:"base_commit_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate checks if the Branch is valid
func (b *Branch) Validate() error {
	if b.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if b.Name == "" {
		return errors.New("name is required")
	}
	if len(b.Name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if b.BaseCommitID == uuid.Nil {
		return errors.New("base_commit_id is required")
	}
	return nil
}

// NewBranch creates a new Branch with generated ID and timestamp
func NewBranch(caseID uuid.UUID, name string, baseCommitID uuid.UUID) *Branch {
	return &Branch{
		ID:           uuid.New(),
		CaseID:       caseID,
		Name:         name,
		BaseCommitID: baseCommitID,
		CreatedAt:    time.Now().UTC(),
	}
}
