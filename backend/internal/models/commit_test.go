package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestCommit_Validate(t *testing.T) {
	validCaseID := uuid.New()

	tests := []struct {
		name    string
		commit  *Commit
		wantErr bool
	}{
		{
			name:    "empty commit should fail",
			commit:  &Commit{},
			wantErr: true,
		},
		{
			name: "missing case_id should fail",
			commit: &Commit{
				ID:      uuid.New(),
				Type:    CommitTypeUploadScan,
				Summary: "Test summary",
			},
			wantErr: true,
		},
		{
			name: "invalid commit type should fail",
			commit: &Commit{
				ID:      uuid.New(),
				CaseID:  validCaseID,
				Type:    CommitType("invalid"),
				Summary: "Test summary",
			},
			wantErr: true,
		},
		{
			name: "missing summary should fail",
			commit: &Commit{
				ID:     uuid.New(),
				CaseID: validCaseID,
				Type:   CommitTypeUploadScan,
			},
			wantErr: true,
		},
		{
			name: "summary too long should fail",
			commit: &Commit{
				ID:      uuid.New(),
				CaseID:  validCaseID,
				Type:    CommitTypeUploadScan,
				Summary: string(make([]byte, 501)),
			},
			wantErr: true,
		},
		{
			name: "valid commit",
			commit: &Commit{
				ID:      uuid.New(),
				CaseID:  validCaseID,
				Type:    CommitTypeUploadScan,
				Summary: "Uploaded 3 scan images",
				Payload: json.RawMessage(`{}`),
			},
			wantErr: false,
		},
		{
			name: "valid commit at max summary length",
			commit: &Commit{
				ID:      uuid.New(),
				CaseID:  validCaseID,
				Type:    CommitTypeReasoningResult,
				Summary: string(make([]byte, 500)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.commit.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Commit.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommit(t *testing.T) {
	caseID := uuid.New()
	commitType := CommitTypeUploadScan
	summary := "Uploaded scan images"
	payload := map[string]interface{}{
		"asset_keys": []string{"key1", "key2"},
	}

	commit, err := NewCommit(caseID, commitType, summary, payload)
	if err != nil {
		t.Fatalf("NewCommit() error = %v", err)
	}

	if commit.ID == uuid.Nil {
		t.Error("NewCommit() should generate a non-nil UUID")
	}

	if commit.CaseID != caseID {
		t.Errorf("NewCommit() CaseID = %v, want %v", commit.CaseID, caseID)
	}

	if commit.Type != commitType {
		t.Errorf("NewCommit() Type = %v, want %v", commit.Type, commitType)
	}

	if commit.Summary != summary {
		t.Errorf("NewCommit() Summary = %v, want %v", commit.Summary, summary)
	}

	if commit.CreatedAt.IsZero() {
		t.Error("NewCommit() should set CreatedAt")
	}

	// Verify payload was serialized
	var decoded map[string]interface{}
	if err := json.Unmarshal(commit.Payload, &decoded); err != nil {
		t.Errorf("NewCommit() payload should be valid JSON: %v", err)
	}

	// Validate the created commit
	if err := commit.Validate(); err != nil {
		t.Errorf("NewCommit() created invalid commit: %v", err)
	}
}

func TestCommit_SetParent(t *testing.T) {
	commit, _ := NewCommit(uuid.New(), CommitTypeUploadScan, "Test", nil)
	parentID := uuid.New()

	commit.SetParent(parentID)

	if commit.ParentCommitID == nil {
		t.Error("SetParent() should set ParentCommitID")
	}

	if *commit.ParentCommitID != parentID {
		t.Errorf("SetParent() ParentCommitID = %v, want %v", *commit.ParentCommitID, parentID)
	}
}

func TestCommit_SetBranch(t *testing.T) {
	commit, _ := NewCommit(uuid.New(), CommitTypeUploadScan, "Test", nil)
	branchID := uuid.New()

	commit.SetBranch(branchID)

	if commit.BranchID == nil {
		t.Error("SetBranch() should set BranchID")
	}

	if *commit.BranchID != branchID {
		t.Errorf("SetBranch() BranchID = %v, want %v", *commit.BranchID, branchID)
	}
}

func TestBranch_Validate(t *testing.T) {
	validCaseID := uuid.New()
	validCommitID := uuid.New()

	tests := []struct {
		name    string
		branch  *Branch
		wantErr bool
	}{
		{
			name:    "empty branch should fail",
			branch:  &Branch{},
			wantErr: true,
		},
		{
			name: "missing case_id should fail",
			branch: &Branch{
				ID:           uuid.New(),
				Name:         "Hypothesis A",
				BaseCommitID: validCommitID,
			},
			wantErr: true,
		},
		{
			name: "missing name should fail",
			branch: &Branch{
				ID:           uuid.New(),
				CaseID:       validCaseID,
				BaseCommitID: validCommitID,
			},
			wantErr: true,
		},
		{
			name: "name too long should fail",
			branch: &Branch{
				ID:           uuid.New(),
				CaseID:       validCaseID,
				Name:         string(make([]byte, 101)),
				BaseCommitID: validCommitID,
			},
			wantErr: true,
		},
		{
			name: "missing base_commit_id should fail",
			branch: &Branch{
				ID:     uuid.New(),
				CaseID: validCaseID,
				Name:   "Hypothesis A",
			},
			wantErr: true,
		},
		{
			name: "valid branch",
			branch: &Branch{
				ID:           uuid.New(),
				CaseID:       validCaseID,
				Name:         "Hypothesis A",
				BaseCommitID: validCommitID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.branch.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Branch.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBranch(t *testing.T) {
	caseID := uuid.New()
	name := "Hypothesis A"
	baseCommitID := uuid.New()

	branch := NewBranch(caseID, name, baseCommitID)

	if branch.ID == uuid.Nil {
		t.Error("NewBranch() should generate a non-nil UUID")
	}

	if branch.CaseID != caseID {
		t.Errorf("NewBranch() CaseID = %v, want %v", branch.CaseID, caseID)
	}

	if branch.Name != name {
		t.Errorf("NewBranch() Name = %v, want %v", branch.Name, name)
	}

	if branch.BaseCommitID != baseCommitID {
		t.Errorf("NewBranch() BaseCommitID = %v, want %v", branch.BaseCommitID, baseCommitID)
	}

	if branch.CreatedAt.IsZero() {
		t.Error("NewBranch() should set CreatedAt")
	}

	if err := branch.Validate(); err != nil {
		t.Errorf("NewBranch() created invalid branch: %v", err)
	}
}
