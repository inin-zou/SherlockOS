package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Asset represents a stored file (image, mesh, etc.)
type Asset struct {
	ID         uuid.UUID              `json:"id"`
	CaseID     uuid.UUID              `json:"case_id"`
	Kind       AssetKind              `json:"kind"`
	StorageKey string                 `json:"storage_key"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Validate checks if the Asset is valid
func (a *Asset) Validate() error {
	if a.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if !a.Kind.IsValid() {
		return errors.New("invalid asset kind")
	}
	if a.StorageKey == "" {
		return errors.New("storage_key is required")
	}
	return nil
}

// NewAsset creates a new Asset with generated ID and timestamp
func NewAsset(caseID uuid.UUID, kind AssetKind, storageKey string) *Asset {
	return &Asset{
		ID:         uuid.New(),
		CaseID:     caseID,
		Kind:       kind,
		StorageKey: storageKey,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now().UTC(),
	}
}

// SetMetadata adds metadata to the asset
func (a *Asset) SetMetadata(key string, value interface{}) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
}

// SceneSnapshot represents the current state of a scene
type SceneSnapshot struct {
	CaseID     uuid.UUID   `json:"case_id"`
	CommitID   uuid.UUID   `json:"commit_id"`
	Scenegraph *SceneGraph `json:"scenegraph"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Validate checks if the SceneSnapshot is valid
func (ss *SceneSnapshot) Validate() error {
	if ss.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if ss.CommitID == uuid.Nil {
		return errors.New("commit_id is required")
	}
	if ss.Scenegraph == nil {
		return errors.New("scenegraph is required")
	}
	return ss.Scenegraph.Validate()
}

// NewSceneSnapshot creates a new SceneSnapshot
func NewSceneSnapshot(caseID, commitID uuid.UUID, scenegraph *SceneGraph) *SceneSnapshot {
	return &SceneSnapshot{
		CaseID:     caseID,
		CommitID:   commitID,
		Scenegraph: scenegraph,
		UpdatedAt:  time.Now().UTC(),
	}
}
