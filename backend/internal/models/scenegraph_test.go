package models

import (
	"encoding/json"
	"testing"
)

func TestSceneGraph_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sg      *SceneGraph
		wantErr bool
	}{
		{
			name:    "empty scenegraph should fail",
			sg:      &SceneGraph{},
			wantErr: true,
		},
		{
			name:    "valid empty scenegraph",
			sg:      NewEmptySceneGraph(),
			wantErr: false,
		},
		{
			name: "invalid bounding box",
			sg: &SceneGraph{
				Version: "1.0.0",
				Bounds: BoundingBox{
					Min: [3]float64{10, 10, 10},
					Max: [3]float64{0, 0, 0}, // max < min
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SceneGraph.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSceneObject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *SceneObject
		wantErr bool
	}{
		{
			name:    "empty object should fail",
			obj:     &SceneObject{},
			wantErr: true,
		},
		{
			name: "missing label should fail",
			obj: &SceneObject{
				ID:   "obj_001",
				Type: ObjectTypeFurniture,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence should fail",
			obj: &SceneObject{
				ID:         "obj_001",
				Type:       ObjectTypeFurniture,
				Label:      "Table",
				State:      ObjectStateVisible,
				Confidence: 1.5, // > 1
			},
			wantErr: true,
		},
		{
			name: "valid object",
			obj: &SceneObject{
				ID:         "obj_001",
				Type:       ObjectTypeFurniture,
				Label:      "Table",
				State:      ObjectStateVisible,
				Confidence: 0.85,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.obj.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SceneObject.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoundingBox_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bbox    BoundingBox
		wantErr bool
	}{
		{
			name: "valid bounding box",
			bbox: BoundingBox{
				Min: [3]float64{0, 0, 0},
				Max: [3]float64{10, 5, 10},
			},
			wantErr: false,
		},
		{
			name: "min equals max (valid)",
			bbox: BoundingBox{
				Min: [3]float64{5, 5, 5},
				Max: [3]float64{5, 5, 5},
			},
			wantErr: false,
		},
		{
			name: "min greater than max on x",
			bbox: BoundingBox{
				Min: [3]float64{10, 0, 0},
				Max: [3]float64{0, 5, 5},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bbox.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BoundingBox.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoundingBox_Contains(t *testing.T) {
	bbox := BoundingBox{
		Min: [3]float64{0, 0, 0},
		Max: [3]float64{10, 10, 10},
	}

	tests := []struct {
		name     string
		point    [3]float64
		expected bool
	}{
		{
			name:     "point inside",
			point:    [3]float64{5, 5, 5},
			expected: true,
		},
		{
			name:     "point on boundary",
			point:    [3]float64{0, 0, 0},
			expected: true,
		},
		{
			name:     "point outside (x too large)",
			point:    [3]float64{15, 5, 5},
			expected: false,
		},
		{
			name:     "point outside (negative)",
			point:    [3]float64{-1, 5, 5},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbox.Contains(tt.point); got != tt.expected {
				t.Errorf("BoundingBox.Contains() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestSceneGraph_JSONRoundTrip(t *testing.T) {
	sg := NewEmptySceneGraph()
	sg.Objects = append(sg.Objects, SceneObject{
		ID:         "obj_001",
		Type:       ObjectTypeDoor,
		Label:      "Main Door",
		Pose:       NewDefaultPose(),
		BBox:       BoundingBox{Min: [3]float64{0, 0, 0}, Max: [3]float64{1, 2, 0.1}},
		State:      ObjectStateVisible,
		Confidence: 0.92,
	})

	// Marshal to JSON
	data, err := json.Marshal(sg)
	if err != nil {
		t.Fatalf("Failed to marshal SceneGraph: %v", err)
	}

	// Unmarshal back
	var sg2 SceneGraph
	if err := json.Unmarshal(data, &sg2); err != nil {
		t.Fatalf("Failed to unmarshal SceneGraph: %v", err)
	}

	// Verify
	if sg2.Version != sg.Version {
		t.Errorf("Version mismatch: got %s, want %s", sg2.Version, sg.Version)
	}
	if len(sg2.Objects) != len(sg.Objects) {
		t.Errorf("Objects count mismatch: got %d, want %d", len(sg2.Objects), len(sg.Objects))
	}
	if sg2.Objects[0].ID != sg.Objects[0].ID {
		t.Errorf("Object ID mismatch: got %s, want %s", sg2.Objects[0].ID, sg.Objects[0].ID)
	}
}

func TestEvidenceCard_Validate(t *testing.T) {
	tests := []struct {
		name    string
		card    *EvidenceCard
		wantErr bool
	}{
		{
			name:    "empty card should fail",
			card:    &EvidenceCard{},
			wantErr: true,
		},
		{
			name: "missing title should fail",
			card: &EvidenceCard{
				ID:         "ev_001",
				Confidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence should fail",
			card: &EvidenceCard{
				ID:         "ev_001",
				Title:      "Evidence",
				Confidence: -0.5,
			},
			wantErr: true,
		},
		{
			name: "valid card",
			card: &EvidenceCard{
				ID:         "ev_001",
				Title:      "Door Lock Damage",
				Confidence: 0.85,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EvidenceCard.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
