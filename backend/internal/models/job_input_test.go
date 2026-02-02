package models

import (
	"testing"
)

func TestReconstructionInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ReconstructionInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: ReconstructionInput{
				CaseID:        "case_123",
				ScanAssetKeys: []string{"key1", "key2"},
			},
			wantErr: false,
		},
		{
			name: "missing case_id",
			input: ReconstructionInput{
				ScanAssetKeys: []string{"key1"},
			},
			wantErr: true,
		},
		{
			name: "empty scan_asset_keys",
			input: ReconstructionInput{
				CaseID:        "case_123",
				ScanAssetKeys: []string{},
			},
			wantErr: true,
		},
		{
			name: "nil scan_asset_keys",
			input: ReconstructionInput{
				CaseID: "case_123",
			},
			wantErr: true,
		},
		{
			name: "empty key in scan_asset_keys",
			input: ReconstructionInput{
				CaseID:        "case_123",
				ScanAssetKeys: []string{"key1", ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReasoningInput_Validate(t *testing.T) {
	validSG := NewEmptySceneGraph()

	tests := []struct {
		name    string
		input   ReasoningInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: ReasoningInput{
				CaseID:         "case_123",
				Scenegraph:     validSG,
				ThinkingBudget: 8192,
				MaxTrajectories: 3,
			},
			wantErr: false,
		},
		{
			name: "missing case_id",
			input: ReasoningInput{
				Scenegraph: validSG,
			},
			wantErr: true,
		},
		{
			name: "missing scenegraph",
			input: ReasoningInput{
				CaseID: "case_123",
			},
			wantErr: true,
		},
		{
			name: "thinking_budget too high",
			input: ReasoningInput{
				CaseID:         "case_123",
				Scenegraph:     validSG,
				ThinkingBudget: 30000,
			},
			wantErr: true,
		},
		{
			name: "thinking_budget negative",
			input: ReasoningInput{
				CaseID:         "case_123",
				Scenegraph:     validSG,
				ThinkingBudget: -1,
			},
			wantErr: true,
		},
		{
			name: "max_trajectories negative",
			input: ReasoningInput{
				CaseID:          "case_123",
				Scenegraph:      validSG,
				MaxTrajectories: -1,
			},
			wantErr: true,
		},
		{
			name: "zero thinking_budget is valid (uses default)",
			input: ReasoningInput{
				CaseID:     "case_123",
				Scenegraph: validSG,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReasoningInput_SetDefaults(t *testing.T) {
	input := ReasoningInput{
		CaseID:     "case_123",
		Scenegraph: NewEmptySceneGraph(),
	}

	input.SetDefaults()

	if input.ThinkingBudget != 8192 {
		t.Errorf("SetDefaults() ThinkingBudget = %v, want 8192", input.ThinkingBudget)
	}

	if input.MaxTrajectories != 3 {
		t.Errorf("SetDefaults() MaxTrajectories = %v, want 3", input.MaxTrajectories)
	}

	// Test that non-zero values are not overwritten
	input2 := ReasoningInput{
		CaseID:          "case_123",
		Scenegraph:      NewEmptySceneGraph(),
		ThinkingBudget:  4096,
		MaxTrajectories: 5,
	}

	input2.SetDefaults()

	if input2.ThinkingBudget != 4096 {
		t.Errorf("SetDefaults() should not overwrite ThinkingBudget, got %v", input2.ThinkingBudget)
	}

	if input2.MaxTrajectories != 5 {
		t.Errorf("SetDefaults() should not overwrite MaxTrajectories, got %v", input2.MaxTrajectories)
	}
}

func TestImageGenInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ImageGenInput
		wantErr bool
	}{
		{
			name: "valid portrait input",
			input: ImageGenInput{
				CaseID:        "case_123",
				GenType:       ImageGenTypePortrait,
				Resolution:    "1k",
				PortraitAttrs: NewEmptySuspectAttributes(),
			},
			wantErr: false,
		},
		{
			name: "valid evidence board input",
			input: ImageGenInput{
				CaseID:     "case_123",
				GenType:    ImageGenTypeEvidenceBoard,
				Resolution: "2k",
			},
			wantErr: false,
		},
		{
			name: "missing case_id",
			input: ImageGenInput{
				GenType:    ImageGenTypePortrait,
				Resolution: "1k",
			},
			wantErr: true,
		},
		{
			name: "invalid gen_type",
			input: ImageGenInput{
				CaseID:     "case_123",
				GenType:    "invalid",
				Resolution: "1k",
			},
			wantErr: true,
		},
		{
			name: "invalid resolution",
			input: ImageGenInput{
				CaseID:     "case_123",
				GenType:    ImageGenTypeEvidenceBoard,
				Resolution: "5k",
			},
			wantErr: true,
		},
		{
			name: "portrait without attributes",
			input: ImageGenInput{
				CaseID:     "case_123",
				GenType:    ImageGenTypePortrait,
				Resolution: "1k",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageGenInput_GetModelForResolution(t *testing.T) {
	tests := []struct {
		resolution string
		expected   string
	}{
		{"1k", "gemini-2.5-flash-image"},      // Nano Banana
		{"2k", "gemini-3-pro-image-preview"},  // Nano Banana Pro
		{"4k", "gemini-3-pro-image-preview"},  // Nano Banana Pro
	}

	for _, tt := range tests {
		t.Run(tt.resolution, func(t *testing.T) {
			input := ImageGenInput{Resolution: tt.resolution}
			result := input.GetModelForResolution()
			if result != tt.expected {
				t.Errorf("GetModelForResolution() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestImageGenType_IsValid(t *testing.T) {
	tests := []struct {
		genType ImageGenType
		want    bool
	}{
		{ImageGenTypePortrait, true},
		{ImageGenTypeEvidenceBoard, true},
		{ImageGenTypeComparison, true},
		{ImageGenTypeReportFigure, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.genType), func(t *testing.T) {
			if got := tt.genType.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileInput_Validate(t *testing.T) {
	validStatement := WitnessStatementInput{
		SourceName:  "Witness A",
		Content:     "The suspect was tall",
		Credibility: 0.8,
	}

	tests := []struct {
		name    string
		input   ProfileInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: ProfileInput{
				CaseID:     "case_123",
				Statements: []WitnessStatementInput{validStatement},
			},
			wantErr: false,
		},
		{
			name: "missing case_id",
			input: ProfileInput{
				Statements: []WitnessStatementInput{validStatement},
			},
			wantErr: true,
		},
		{
			name: "empty statements",
			input: ProfileInput{
				CaseID:     "case_123",
				Statements: []WitnessStatementInput{},
			},
			wantErr: true,
		},
		{
			name: "invalid statement in list",
			input: ProfileInput{
				CaseID: "case_123",
				Statements: []WitnessStatementInput{
					{SourceName: "", Content: "Test", Credibility: 0.5}, // Missing source name
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
