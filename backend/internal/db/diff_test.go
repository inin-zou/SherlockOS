package db

import (
	"testing"

	"github.com/sherlockos/backend/internal/models"
)

func TestComputeSceneGraphDiff_NilInputs(t *testing.T) {
	// Both nil
	diff := ComputeSceneGraphDiff(nil, nil)
	if diff == nil {
		t.Fatal("diff should not be nil")
	}
	if len(diff.ObjectsAdded) != 0 {
		t.Errorf("ObjectsAdded = %d, want 0", len(diff.ObjectsAdded))
	}
	if len(diff.ObjectsRemoved) != 0 {
		t.Errorf("ObjectsRemoved = %d, want 0", len(diff.ObjectsRemoved))
	}
}

func TestComputeSceneGraphDiff_AddObjects(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	// Add objects to 'to'
	to.Objects = []models.SceneObject{
		{
			ID:         "obj-1",
			Type:       models.ObjectTypePersonMarker,
			Label:      "Suspect",
			State:      models.ObjectStateVisible,
			Confidence: 0.9,
		},
		{
			ID:         "obj-2",
			Type:       models.ObjectTypeWeapon,
			Label:      "Knife",
			State:      models.ObjectStateSuspicious,
			Confidence: 0.7,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.ObjectsAdded) != 2 {
		t.Errorf("ObjectsAdded = %d, want 2", len(diff.ObjectsAdded))
	}
	if len(diff.ObjectsRemoved) != 0 {
		t.Errorf("ObjectsRemoved = %d, want 0", len(diff.ObjectsRemoved))
	}
	if len(diff.ObjectsUpdated) != 0 {
		t.Errorf("ObjectsUpdated = %d, want 0", len(diff.ObjectsUpdated))
	}
}

func TestComputeSceneGraphDiff_RemoveObjects(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	// Add objects to 'from'
	from.Objects = []models.SceneObject{
		{
			ID:         "obj-1",
			Type:       models.ObjectTypePersonMarker,
			Label:      "Suspect",
			State:      models.ObjectStateVisible,
			Confidence: 0.9,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.ObjectsAdded) != 0 {
		t.Errorf("ObjectsAdded = %d, want 0", len(diff.ObjectsAdded))
	}
	if len(diff.ObjectsRemoved) != 1 {
		t.Errorf("ObjectsRemoved = %d, want 1", len(diff.ObjectsRemoved))
	}
	if diff.ObjectsRemoved[0] != "obj-1" {
		t.Errorf("ObjectsRemoved[0] = %s, want obj-1", diff.ObjectsRemoved[0])
	}
}

func TestComputeSceneGraphDiff_UpdateObjects(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	// Same object with different label
	from.Objects = []models.SceneObject{
		{
			ID:         "obj-1",
			Type:       models.ObjectTypePersonMarker,
			Label:      "Unknown Person",
			State:      models.ObjectStateSuspicious,
			Confidence: 0.5,
		},
	}

	to.Objects = []models.SceneObject{
		{
			ID:         "obj-1",
			Type:       models.ObjectTypePersonMarker,
			Label:      "John Doe",
			State:      models.ObjectStateVisible,
			Confidence: 0.9,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.ObjectsAdded) != 0 {
		t.Errorf("ObjectsAdded = %d, want 0", len(diff.ObjectsAdded))
	}
	if len(diff.ObjectsRemoved) != 0 {
		t.Errorf("ObjectsRemoved = %d, want 0", len(diff.ObjectsRemoved))
	}
	if len(diff.ObjectsUpdated) != 1 {
		t.Errorf("ObjectsUpdated = %d, want 1", len(diff.ObjectsUpdated))
	}

	update := diff.ObjectsUpdated[0]
	if update.ID != "obj-1" {
		t.Errorf("update.ID = %s, want obj-1", update.ID)
	}
	if update.Before.Label != "Unknown Person" {
		t.Errorf("update.Before.Label = %s, want Unknown Person", update.Before.Label)
	}
	if update.After.Label != "John Doe" {
		t.Errorf("update.After.Label = %s, want John Doe", update.After.Label)
	}
}

func TestComputeSceneGraphDiff_AddEvidence(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	to.Evidence = []models.EvidenceCard{
		{
			ID:          "ev-1",
			Title:       "Blood sample",
			Description: "DNA match found",
			Confidence:  0.95,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.EvidenceAdded) != 1 {
		t.Errorf("EvidenceAdded = %d, want 1", len(diff.EvidenceAdded))
	}
	if diff.EvidenceAdded[0].ID != "ev-1" {
		t.Errorf("EvidenceAdded[0].ID = %s, want ev-1", diff.EvidenceAdded[0].ID)
	}
}

func TestComputeSceneGraphDiff_RemoveEvidence(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	from.Evidence = []models.EvidenceCard{
		{
			ID:          "ev-1",
			Title:       "Fingerprint",
			Description: "Partial print",
			Confidence:  0.6,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.EvidenceRemoved) != 1 {
		t.Errorf("EvidenceRemoved = %d, want 1", len(diff.EvidenceRemoved))
	}
	if diff.EvidenceRemoved[0] != "ev-1" {
		t.Errorf("EvidenceRemoved[0] = %s, want ev-1", diff.EvidenceRemoved[0])
	}
}

func TestComputeSceneGraphDiff_UpdateEvidence(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	from.Evidence = []models.EvidenceCard{
		{
			ID:          "ev-1",
			Title:       "Fingerprint",
			Description: "Partial print",
			Confidence:  0.6,
		},
	}

	to.Evidence = []models.EvidenceCard{
		{
			ID:          "ev-1",
			Title:       "Fingerprint",
			Description: "Full match to suspect",
			Confidence:  0.95,
		},
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.EvidenceUpdated) != 1 {
		t.Errorf("EvidenceUpdated = %d, want 1", len(diff.EvidenceUpdated))
	}
	if diff.EvidenceUpdated[0].Before.Confidence != 0.6 {
		t.Errorf("EvidenceUpdated[0].Before.Confidence = %f, want 0.6", diff.EvidenceUpdated[0].Before.Confidence)
	}
	if diff.EvidenceUpdated[0].After.Confidence != 0.95 {
		t.Errorf("EvidenceUpdated[0].After.Confidence = %f, want 0.95", diff.EvidenceUpdated[0].After.Confidence)
	}
}

func TestComputeSceneGraphDiff_MixedChanges(t *testing.T) {
	from := models.NewEmptySceneGraph()
	to := models.NewEmptySceneGraph()

	// From has 3 objects: obj-1 (will be removed), obj-2 (will be updated), obj-3 (unchanged)
	from.Objects = []models.SceneObject{
		{ID: "obj-1", Type: models.ObjectTypePersonMarker, Label: "Person A", State: models.ObjectStateVisible, Confidence: 0.9},
		{ID: "obj-2", Type: models.ObjectTypeWeapon, Label: "Gun", State: models.ObjectStateSuspicious, Confidence: 0.5},
		{ID: "obj-3", Type: models.ObjectTypeFurniture, Label: "Table", State: models.ObjectStateVisible, Confidence: 0.99},
	}

	// To has: obj-2 (updated), obj-3 (unchanged), obj-4 (added)
	to.Objects = []models.SceneObject{
		{ID: "obj-2", Type: models.ObjectTypeWeapon, Label: "Knife", State: models.ObjectStateVisible, Confidence: 0.8}, // Updated
		{ID: "obj-3", Type: models.ObjectTypeFurniture, Label: "Table", State: models.ObjectStateVisible, Confidence: 0.99}, // Unchanged
		{ID: "obj-4", Type: models.ObjectTypePersonMarker, Label: "Person B", State: models.ObjectStateSuspicious, Confidence: 0.7}, // Added
	}

	diff := ComputeSceneGraphDiff(from, to)

	if len(diff.ObjectsAdded) != 1 {
		t.Errorf("ObjectsAdded = %d, want 1", len(diff.ObjectsAdded))
	}
	if len(diff.ObjectsRemoved) != 1 {
		t.Errorf("ObjectsRemoved = %d, want 1", len(diff.ObjectsRemoved))
	}
	if len(diff.ObjectsUpdated) != 1 {
		t.Errorf("ObjectsUpdated = %d, want 1", len(diff.ObjectsUpdated))
	}

	// Verify added object
	if diff.ObjectsAdded[0].ID != "obj-4" {
		t.Errorf("ObjectsAdded[0].ID = %s, want obj-4", diff.ObjectsAdded[0].ID)
	}

	// Verify removed object
	if diff.ObjectsRemoved[0] != "obj-1" {
		t.Errorf("ObjectsRemoved[0] = %s, want obj-1", diff.ObjectsRemoved[0])
	}

	// Verify updated object
	if diff.ObjectsUpdated[0].ID != "obj-2" {
		t.Errorf("ObjectsUpdated[0].ID = %s, want obj-2", diff.ObjectsUpdated[0].ID)
	}
}

func TestObjectsEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        models.SceneObject
		b        models.SceneObject
		expected bool
	}{
		{
			name: "identical objects",
			a: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			b: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			expected: true,
		},
		{
			name: "different label",
			a: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person A",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			b: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person B",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			expected: false,
		},
		{
			name: "different confidence",
			a: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.5,
				Pose:       models.NewDefaultPose(),
			},
			b: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			expected: false,
		},
		{
			name: "different state",
			a: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateSuspicious,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			b: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose:       models.NewDefaultPose(),
			},
			expected: false,
		},
		{
			name: "different position",
			a: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose: models.Pose{
					Position: [3]float64{0, 0, 0},
					Rotation: [4]float64{1, 0, 0, 0},
				},
			},
			b: models.SceneObject{
				ID:         "obj-1",
				Type:       models.ObjectTypePersonMarker,
				Label:      "Person",
				State:      models.ObjectStateVisible,
				Confidence: 0.9,
				Pose: models.Pose{
					Position: [3]float64{1, 0, 0},
					Rotation: [4]float64{1, 0, 0, 0},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := objectsEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("objectsEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvidenceEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        models.EvidenceCard
		b        models.EvidenceCard
		expected bool
	}{
		{
			name: "identical evidence",
			a: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Blood Sample",
				Description: "DNA match",
				Confidence:  0.95,
			},
			b: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Blood Sample",
				Description: "DNA match",
				Confidence:  0.95,
			},
			expected: true,
		},
		{
			name: "different title",
			a: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Blood Sample",
				Description: "DNA match",
				Confidence:  0.95,
			},
			b: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Hair Sample",
				Description: "DNA match",
				Confidence:  0.95,
			},
			expected: false,
		},
		{
			name: "different confidence",
			a: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Blood Sample",
				Description: "DNA match",
				Confidence:  0.5,
			},
			b: models.EvidenceCard{
				ID:          "ev-1",
				Title:       "Blood Sample",
				Description: "DNA match",
				Confidence:  0.95,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evidenceEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("evidenceEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSceneGraphDiff_EmptyInitialization(t *testing.T) {
	diff := ComputeSceneGraphDiff(nil, nil)

	// Verify all slices are initialized (not nil)
	if diff.ObjectsAdded == nil {
		t.Error("ObjectsAdded should not be nil")
	}
	if diff.ObjectsUpdated == nil {
		t.Error("ObjectsUpdated should not be nil")
	}
	if diff.ObjectsRemoved == nil {
		t.Error("ObjectsRemoved should not be nil")
	}
	if diff.EvidenceAdded == nil {
		t.Error("EvidenceAdded should not be nil")
	}
	if diff.EvidenceUpdated == nil {
		t.Error("EvidenceUpdated should not be nil")
	}
	if diff.EvidenceRemoved == nil {
		t.Error("EvidenceRemoved should not be nil")
	}
}
