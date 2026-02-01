package models

// CommitType represents the type of a commit
type CommitType string

const (
	CommitTypeUploadScan           CommitType = "upload_scan"
	CommitTypeWitnessStatement     CommitType = "witness_statement"
	CommitTypeManualEdit           CommitType = "manual_edit"
	CommitTypeReconstructionUpdate CommitType = "reconstruction_update"
	CommitTypeProfileUpdate        CommitType = "profile_update"
	CommitTypeReasoningResult      CommitType = "reasoning_result"
	CommitTypeExportReport         CommitType = "export_report"
)

// IsValid checks if the commit type is valid
func (ct CommitType) IsValid() bool {
	switch ct {
	case CommitTypeUploadScan, CommitTypeWitnessStatement, CommitTypeManualEdit,
		CommitTypeReconstructionUpdate, CommitTypeProfileUpdate,
		CommitTypeReasoningResult, CommitTypeExportReport:
		return true
	}
	return false
}

// JobType represents the type of an async job
type JobType string

const (
	JobTypeReconstruction JobType = "reconstruction"
	JobTypeImageGen       JobType = "imagegen"
	JobTypeReasoning      JobType = "reasoning"
	JobTypeProfile        JobType = "profile"
	JobTypeExport         JobType = "export"
)

// IsValid checks if the job type is valid
func (jt JobType) IsValid() bool {
	switch jt {
	case JobTypeReconstruction, JobTypeImageGen, JobTypeReasoning, JobTypeProfile, JobTypeExport:
		return true
	}
	return false
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued   JobStatus = "queued"
	JobStatusRunning  JobStatus = "running"
	JobStatusDone     JobStatus = "done"
	JobStatusFailed   JobStatus = "failed"
	JobStatusCanceled JobStatus = "canceled"
)

// IsValid checks if the job status is valid
func (js JobStatus) IsValid() bool {
	switch js {
	case JobStatusQueued, JobStatusRunning, JobStatusDone, JobStatusFailed, JobStatusCanceled:
		return true
	}
	return false
}

// IsTerminal returns true if the status is a terminal state
func (js JobStatus) IsTerminal() bool {
	return js == JobStatusDone || js == JobStatusFailed || js == JobStatusCanceled
}

// AssetKind represents the type of an asset
type AssetKind string

const (
	AssetKindScanImage      AssetKind = "scan_image"
	AssetKindGeneratedImage AssetKind = "generated_image"
	AssetKindMesh           AssetKind = "mesh"
	AssetKindPointcloud     AssetKind = "pointcloud"
	AssetKindPortrait       AssetKind = "portrait"
	AssetKindReport         AssetKind = "report"
)

// IsValid checks if the asset kind is valid
func (ak AssetKind) IsValid() bool {
	switch ak {
	case AssetKindScanImage, AssetKindGeneratedImage, AssetKindMesh,
		AssetKindPointcloud, AssetKindPortrait, AssetKindReport:
		return true
	}
	return false
}

// ObjectType represents the type of a scene object
type ObjectType string

const (
	ObjectTypeFurniture    ObjectType = "furniture"
	ObjectTypeDoor         ObjectType = "door"
	ObjectTypeWindow       ObjectType = "window"
	ObjectTypeWall         ObjectType = "wall"
	ObjectTypeEvidenceItem ObjectType = "evidence_item"
	ObjectTypeWeapon       ObjectType = "weapon"
	ObjectTypeFootprint    ObjectType = "footprint"
	ObjectTypeBloodstain   ObjectType = "bloodstain"
	ObjectTypeVehicle      ObjectType = "vehicle"
	ObjectTypePersonMarker ObjectType = "person_marker"
	ObjectTypeOther        ObjectType = "other"
)

// IsValid checks if the object type is valid
func (ot ObjectType) IsValid() bool {
	switch ot {
	case ObjectTypeFurniture, ObjectTypeDoor, ObjectTypeWindow, ObjectTypeWall,
		ObjectTypeEvidenceItem, ObjectTypeWeapon, ObjectTypeFootprint,
		ObjectTypeBloodstain, ObjectTypeVehicle, ObjectTypePersonMarker, ObjectTypeOther:
		return true
	}
	return false
}

// ObjectState represents the visibility state of an object
type ObjectState string

const (
	ObjectStateVisible    ObjectState = "visible"
	ObjectStateOccluded   ObjectState = "occluded"
	ObjectStateSuspicious ObjectState = "suspicious"
	ObjectStateRemoved    ObjectState = "removed"
)

// IsValid checks if the object state is valid
func (os ObjectState) IsValid() bool {
	switch os {
	case ObjectStateVisible, ObjectStateOccluded, ObjectStateSuspicious, ObjectStateRemoved:
		return true
	}
	return false
}

// ConstraintType represents the type of a constraint
type ConstraintType string

const (
	ConstraintTypeDoorDirection ConstraintType = "door_direction"
	ConstraintTypePassableArea  ConstraintType = "passable_area"
	ConstraintTypeHeightRange   ConstraintType = "height_range"
	ConstraintTypeTimeWindow    ConstraintType = "time_window"
	ConstraintTypeCustom        ConstraintType = "custom"
)

// IsValid checks if the constraint type is valid
func (ct ConstraintType) IsValid() bool {
	switch ct {
	case ConstraintTypeDoorDirection, ConstraintTypePassableArea,
		ConstraintTypeHeightRange, ConstraintTypeTimeWindow, ConstraintTypeCustom:
		return true
	}
	return false
}

// UncertaintyLevel represents the level of uncertainty
type UncertaintyLevel string

const (
	UncertaintyLevelLow    UncertaintyLevel = "low"
	UncertaintyLevelMedium UncertaintyLevel = "medium"
	UncertaintyLevelHigh   UncertaintyLevel = "high"
)

// IsValid checks if the uncertainty level is valid
func (ul UncertaintyLevel) IsValid() bool {
	switch ul {
	case UncertaintyLevelLow, UncertaintyLevelMedium, UncertaintyLevelHigh:
		return true
	}
	return false
}

// EvidenceSourceType represents the source type of evidence
type EvidenceSourceType string

const (
	EvidenceSourceTypeUpload    EvidenceSourceType = "upload"
	EvidenceSourceTypeWitness   EvidenceSourceType = "witness"
	EvidenceSourceTypeInference EvidenceSourceType = "inference"
)

// IsValid checks if the evidence source type is valid
func (est EvidenceSourceType) IsValid() bool {
	switch est {
	case EvidenceSourceTypeUpload, EvidenceSourceTypeWitness, EvidenceSourceTypeInference:
		return true
	}
	return false
}
