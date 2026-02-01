package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestAsset_Validate(t *testing.T) {
	validCaseID := uuid.New()

	tests := []struct {
		name    string
		asset   *Asset
		wantErr bool
	}{
		{
			name:    "empty asset should fail",
			asset:   &Asset{},
			wantErr: true,
		},
		{
			name: "missing case_id should fail",
			asset: &Asset{
				ID:         uuid.New(),
				Kind:       AssetKindScanImage,
				StorageKey: "cases/123/scans/image.jpg",
			},
			wantErr: true,
		},
		{
			name: "invalid kind should fail",
			asset: &Asset{
				ID:         uuid.New(),
				CaseID:     validCaseID,
				Kind:       AssetKind("invalid"),
				StorageKey: "cases/123/scans/image.jpg",
			},
			wantErr: true,
		},
		{
			name: "missing storage_key should fail",
			asset: &Asset{
				ID:     uuid.New(),
				CaseID: validCaseID,
				Kind:   AssetKindScanImage,
			},
			wantErr: true,
		},
		{
			name: "valid asset",
			asset: &Asset{
				ID:         uuid.New(),
				CaseID:     validCaseID,
				Kind:       AssetKindScanImage,
				StorageKey: "cases/123/scans/image.jpg",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Asset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewAsset(t *testing.T) {
	caseID := uuid.New()
	kind := AssetKindScanImage
	storageKey := "cases/123/scans/image.jpg"

	asset := NewAsset(caseID, kind, storageKey)

	if asset.ID == uuid.Nil {
		t.Error("NewAsset() should generate a non-nil UUID")
	}

	if asset.CaseID != caseID {
		t.Errorf("NewAsset() CaseID = %v, want %v", asset.CaseID, caseID)
	}

	if asset.Kind != kind {
		t.Errorf("NewAsset() Kind = %v, want %v", asset.Kind, kind)
	}

	if asset.StorageKey != storageKey {
		t.Errorf("NewAsset() StorageKey = %v, want %v", asset.StorageKey, storageKey)
	}

	if asset.Metadata == nil {
		t.Error("NewAsset() should initialize Metadata map")
	}

	if asset.CreatedAt.IsZero() {
		t.Error("NewAsset() should set CreatedAt")
	}

	if err := asset.Validate(); err != nil {
		t.Errorf("NewAsset() created invalid asset: %v", err)
	}
}

func TestAsset_SetMetadata(t *testing.T) {
	asset := NewAsset(uuid.New(), AssetKindScanImage, "key")

	asset.SetMetadata("width", 1920)
	asset.SetMetadata("height", 1080)
	asset.SetMetadata("format", "jpeg")

	if asset.Metadata["width"] != 1920 {
		t.Errorf("SetMetadata() width = %v, want 1920", asset.Metadata["width"])
	}

	if asset.Metadata["height"] != 1080 {
		t.Errorf("SetMetadata() height = %v, want 1080", asset.Metadata["height"])
	}

	if asset.Metadata["format"] != "jpeg" {
		t.Errorf("SetMetadata() format = %v, want jpeg", asset.Metadata["format"])
	}
}

func TestAsset_SetMetadata_NilMap(t *testing.T) {
	asset := &Asset{
		ID:         uuid.New(),
		CaseID:     uuid.New(),
		Kind:       AssetKindScanImage,
		StorageKey: "key",
		Metadata:   nil, // Explicitly nil
	}

	// Should not panic
	asset.SetMetadata("key", "value")

	if asset.Metadata["key"] != "value" {
		t.Error("SetMetadata() should initialize nil map")
	}
}

func TestSceneSnapshot_Validate(t *testing.T) {
	validCaseID := uuid.New()
	validCommitID := uuid.New()

	tests := []struct {
		name     string
		snapshot *SceneSnapshot
		wantErr  bool
	}{
		{
			name:     "empty snapshot should fail",
			snapshot: &SceneSnapshot{},
			wantErr:  true,
		},
		{
			name: "missing case_id should fail",
			snapshot: &SceneSnapshot{
				CommitID:   validCommitID,
				Scenegraph: NewEmptySceneGraph(),
			},
			wantErr: true,
		},
		{
			name: "missing commit_id should fail",
			snapshot: &SceneSnapshot{
				CaseID:     validCaseID,
				Scenegraph: NewEmptySceneGraph(),
			},
			wantErr: true,
		},
		{
			name: "missing scenegraph should fail",
			snapshot: &SceneSnapshot{
				CaseID:   validCaseID,
				CommitID: validCommitID,
			},
			wantErr: true,
		},
		{
			name: "invalid scenegraph should fail",
			snapshot: &SceneSnapshot{
				CaseID:     validCaseID,
				CommitID:   validCommitID,
				Scenegraph: &SceneGraph{}, // Invalid - missing version
			},
			wantErr: true,
		},
		{
			name: "valid snapshot",
			snapshot: &SceneSnapshot{
				CaseID:     validCaseID,
				CommitID:   validCommitID,
				Scenegraph: NewEmptySceneGraph(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.snapshot.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SceneSnapshot.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSceneSnapshot(t *testing.T) {
	caseID := uuid.New()
	commitID := uuid.New()
	sg := NewEmptySceneGraph()

	snapshot := NewSceneSnapshot(caseID, commitID, sg)

	if snapshot.CaseID != caseID {
		t.Errorf("NewSceneSnapshot() CaseID = %v, want %v", snapshot.CaseID, caseID)
	}

	if snapshot.CommitID != commitID {
		t.Errorf("NewSceneSnapshot() CommitID = %v, want %v", snapshot.CommitID, commitID)
	}

	if snapshot.Scenegraph != sg {
		t.Error("NewSceneSnapshot() should set Scenegraph")
	}

	if snapshot.UpdatedAt.IsZero() {
		t.Error("NewSceneSnapshot() should set UpdatedAt")
	}

	if err := snapshot.Validate(); err != nil {
		t.Errorf("NewSceneSnapshot() created invalid snapshot: %v", err)
	}
}
