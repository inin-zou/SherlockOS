package models

import (
	"encoding/json"
	"errors"
)

// SceneGraph represents the world state of a crime scene
type SceneGraph struct {
	Version            string              `json:"version"`
	Bounds             BoundingBox         `json:"bounds"`
	Objects            []SceneObject       `json:"objects"`
	Evidence           []EvidenceCard      `json:"evidence"`
	Constraints        []Constraint        `json:"constraints"`
	UncertaintyRegions []UncertaintyRegion `json:"uncertainty_regions,omitempty"`
}

// Validate checks if the SceneGraph is valid
func (sg *SceneGraph) Validate() error {
	if sg.Version == "" {
		return errors.New("version is required")
	}
	if err := sg.Bounds.Validate(); err != nil {
		return err
	}
	for i, obj := range sg.Objects {
		if err := obj.Validate(); err != nil {
			return errors.New("object " + string(rune(i)) + ": " + err.Error())
		}
	}
	for i, ev := range sg.Evidence {
		if err := ev.Validate(); err != nil {
			return errors.New("evidence " + string(rune(i)) + ": " + err.Error())
		}
	}
	return nil
}

// SceneObject represents an entity in the scene
type SceneObject struct {
	ID              string                 `json:"id"`
	Type            ObjectType             `json:"type"`
	Label           string                 `json:"label"`
	Pose            Pose                   `json:"pose"`
	BBox            BoundingBox            `json:"bbox"`
	MeshRef         string                 `json:"mesh_ref,omitempty"`
	State           ObjectState            `json:"state"`
	EvidenceIDs     []string               `json:"evidence_ids"`
	Confidence      float64                `json:"confidence"`
	SourceCommitIDs []string               `json:"source_commit_ids"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Validate checks if the SceneObject is valid
func (so *SceneObject) Validate() error {
	if so.ID == "" {
		return errors.New("id is required")
	}
	if !so.Type.IsValid() {
		return errors.New("invalid object type")
	}
	if so.Label == "" {
		return errors.New("label is required")
	}
	if !so.State.IsValid() {
		return errors.New("invalid object state")
	}
	if so.Confidence < 0 || so.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	return nil
}

// Pose represents position and orientation in 3D space
type Pose struct {
	Position [3]float64 `json:"position"` // [x, y, z] in meters
	Rotation [4]float64 `json:"rotation"` // quaternion [w, x, y, z]
	Scale    [3]float64 `json:"scale,omitempty"`
}

// NewDefaultPose creates a pose at origin with no rotation
func NewDefaultPose() Pose {
	return Pose{
		Position: [3]float64{0, 0, 0},
		Rotation: [4]float64{1, 0, 0, 0}, // identity quaternion
		Scale:    [3]float64{1, 1, 1},
	}
}

// BoundingBox represents an axis-aligned bounding box
type BoundingBox struct {
	Min [3]float64 `json:"min"`
	Max [3]float64 `json:"max"`
}

// Validate checks if the bounding box is valid (min < max for all axes)
func (bb *BoundingBox) Validate() error {
	for i := 0; i < 3; i++ {
		if bb.Min[i] > bb.Max[i] {
			return errors.New("min must be less than max for all axes")
		}
	}
	return nil
}

// Contains checks if a point is inside the bounding box
func (bb *BoundingBox) Contains(point [3]float64) bool {
	for i := 0; i < 3; i++ {
		if point[i] < bb.Min[i] || point[i] > bb.Max[i] {
			return false
		}
	}
	return true
}

// EvidenceCard represents a piece of evidence
type EvidenceCard struct {
	ID          string           `json:"id"`
	ObjectIDs   []string         `json:"object_ids"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Confidence  float64          `json:"confidence"`
	Sources     []EvidenceSource `json:"sources"`
	Conflicts   []EvidenceSource `json:"conflicts,omitempty"`
	CreatedAt   string           `json:"created_at"`
}

// Validate checks if the EvidenceCard is valid
func (ec *EvidenceCard) Validate() error {
	if ec.ID == "" {
		return errors.New("id is required")
	}
	if ec.Title == "" {
		return errors.New("title is required")
	}
	if ec.Confidence < 0 || ec.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	return nil
}

// EvidenceSource represents the source of evidence
type EvidenceSource struct {
	Type        EvidenceSourceType `json:"type"`
	CommitID    string             `json:"commit_id"`
	Description string             `json:"description,omitempty"`
	Credibility float64            `json:"credibility,omitempty"` // 0-1, only for witness type
}

// Validate checks if the EvidenceSource is valid
func (es *EvidenceSource) Validate() error {
	if !es.Type.IsValid() {
		return errors.New("invalid source type")
	}
	if es.CommitID == "" {
		return errors.New("commit_id is required")
	}
	if es.Type == EvidenceSourceTypeWitness && (es.Credibility < 0 || es.Credibility > 1) {
		return errors.New("credibility must be between 0 and 1 for witness sources")
	}
	return nil
}

// Constraint represents a constraint in the scene
type Constraint struct {
	ID          string                 `json:"id"`
	Type        ConstraintType         `json:"type"`
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Confidence  float64                `json:"confidence"`
}

// Validate checks if the Constraint is valid
func (c *Constraint) Validate() error {
	if c.ID == "" {
		return errors.New("id is required")
	}
	if !c.Type.IsValid() {
		return errors.New("invalid constraint type")
	}
	if c.Confidence < 0 || c.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	return nil
}

// UncertaintyRegion represents an area with uncertain reconstruction
type UncertaintyRegion struct {
	ID     string           `json:"id"`
	BBox   BoundingBox      `json:"bbox"`
	Level  UncertaintyLevel `json:"level"`
	Reason string           `json:"reason"`
}

// Validate checks if the UncertaintyRegion is valid
func (ur *UncertaintyRegion) Validate() error {
	if ur.ID == "" {
		return errors.New("id is required")
	}
	if !ur.Level.IsValid() {
		return errors.New("invalid uncertainty level")
	}
	return nil
}

// NewEmptySceneGraph creates an empty SceneGraph with default values
func NewEmptySceneGraph() *SceneGraph {
	return &SceneGraph{
		Version: "1.0.0",
		Bounds: BoundingBox{
			Min: [3]float64{0, 0, 0},
			Max: [3]float64{10, 3, 10},
		},
		Objects:            []SceneObject{},
		Evidence:           []EvidenceCard{},
		Constraints:        []Constraint{},
		UncertaintyRegions: []UncertaintyRegion{},
	}
}

// MarshalJSON implements custom JSON marshaling
func (sg *SceneGraph) MarshalJSON() ([]byte, error) {
	type Alias SceneGraph
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(sg),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling
func (sg *SceneGraph) UnmarshalJSON(data []byte) error {
	type Alias SceneGraph
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(sg),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
