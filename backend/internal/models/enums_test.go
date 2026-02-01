package models

import "testing"

func TestCommitType_IsValid(t *testing.T) {
	tests := []struct {
		ct   CommitType
		want bool
	}{
		{CommitTypeUploadScan, true},
		{CommitTypeWitnessStatement, true},
		{CommitTypeManualEdit, true},
		{CommitTypeReconstructionUpdate, true},
		{CommitTypeProfileUpdate, true},
		{CommitTypeReasoningResult, true},
		{CommitTypeExportReport, true},
		{CommitType("invalid"), false},
		{CommitType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.ct), func(t *testing.T) {
			if got := tt.ct.IsValid(); got != tt.want {
				t.Errorf("CommitType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobType_IsValid(t *testing.T) {
	tests := []struct {
		jt   JobType
		want bool
	}{
		{JobTypeReconstruction, true},
		{JobTypeImageGen, true},
		{JobTypeReasoning, true},
		{JobTypeProfile, true},
		{JobTypeExport, true},
		{JobType("invalid"), false},
		{JobType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.jt), func(t *testing.T) {
			if got := tt.jt.IsValid(); got != tt.want {
				t.Errorf("JobType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobStatus_IsValid(t *testing.T) {
	tests := []struct {
		js   JobStatus
		want bool
	}{
		{JobStatusQueued, true},
		{JobStatusRunning, true},
		{JobStatusDone, true},
		{JobStatusFailed, true},
		{JobStatusCanceled, true},
		{JobStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.js), func(t *testing.T) {
			if got := tt.js.IsValid(); got != tt.want {
				t.Errorf("JobStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		js   JobStatus
		want bool
	}{
		{JobStatusQueued, false},
		{JobStatusRunning, false},
		{JobStatusDone, true},
		{JobStatusFailed, true},
		{JobStatusCanceled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.js), func(t *testing.T) {
			if got := tt.js.IsTerminal(); got != tt.want {
				t.Errorf("JobStatus.IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectType_IsValid(t *testing.T) {
	validTypes := []ObjectType{
		ObjectTypeFurniture, ObjectTypeDoor, ObjectTypeWindow, ObjectTypeWall,
		ObjectTypeEvidenceItem, ObjectTypeWeapon, ObjectTypeFootprint,
		ObjectTypeBloodstain, ObjectTypeVehicle, ObjectTypePersonMarker, ObjectTypeOther,
	}

	for _, ot := range validTypes {
		t.Run(string(ot), func(t *testing.T) {
			if !ot.IsValid() {
				t.Errorf("ObjectType.IsValid() = false, want true for %s", ot)
			}
		})
	}

	invalid := ObjectType("invalid_type")
	if invalid.IsValid() {
		t.Error("ObjectType.IsValid() = true for invalid type, want false")
	}
}

func TestAssetKind_IsValid(t *testing.T) {
	validKinds := []AssetKind{
		AssetKindScanImage, AssetKindGeneratedImage, AssetKindMesh,
		AssetKindPointcloud, AssetKindPortrait, AssetKindReport,
	}

	for _, ak := range validKinds {
		t.Run(string(ak), func(t *testing.T) {
			if !ak.IsValid() {
				t.Errorf("AssetKind.IsValid() = false, want true for %s", ak)
			}
		})
	}

	invalid := AssetKind("invalid_kind")
	if invalid.IsValid() {
		t.Error("AssetKind.IsValid() = true for invalid kind, want false")
	}
}

func TestConstraintType_IsValid(t *testing.T) {
	validTypes := []ConstraintType{
		ConstraintTypeDoorDirection, ConstraintTypePassableArea,
		ConstraintTypeHeightRange, ConstraintTypeTimeWindow, ConstraintTypeCustom,
	}

	for _, ct := range validTypes {
		t.Run(string(ct), func(t *testing.T) {
			if !ct.IsValid() {
				t.Errorf("ConstraintType.IsValid() = false, want true for %s", ct)
			}
		})
	}
}

func TestUncertaintyLevel_IsValid(t *testing.T) {
	tests := []struct {
		ul   UncertaintyLevel
		want bool
	}{
		{UncertaintyLevelLow, true},
		{UncertaintyLevelMedium, true},
		{UncertaintyLevelHigh, true},
		{UncertaintyLevel("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.ul), func(t *testing.T) {
			if got := tt.ul.IsValid(); got != tt.want {
				t.Errorf("UncertaintyLevel.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
