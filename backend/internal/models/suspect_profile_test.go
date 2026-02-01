package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestSuspectProfile_Validate(t *testing.T) {
	validCaseID := uuid.New()
	validCommitID := uuid.New()

	tests := []struct {
		name    string
		profile *SuspectProfile
		wantErr bool
	}{
		{
			name:    "empty profile should fail",
			profile: &SuspectProfile{},
			wantErr: true,
		},
		{
			name: "missing case_id should fail",
			profile: &SuspectProfile{
				CommitID: validCommitID,
			},
			wantErr: true,
		},
		{
			name: "missing commit_id should fail",
			profile: &SuspectProfile{
				CaseID: validCaseID,
			},
			wantErr: true,
		},
		{
			name: "valid profile without attributes",
			profile: &SuspectProfile{
				CaseID:   validCaseID,
				CommitID: validCommitID,
			},
			wantErr: false,
		},
		{
			name: "valid profile with attributes",
			profile: &SuspectProfile{
				CaseID:     validCaseID,
				CommitID:   validCommitID,
				Attributes: NewEmptySuspectAttributes(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SuspectProfile.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSuspectProfile(t *testing.T) {
	caseID := uuid.New()
	commitID := uuid.New()

	profile := NewSuspectProfile(caseID, commitID)

	if profile.CaseID != caseID {
		t.Errorf("NewSuspectProfile() CaseID = %v, want %v", profile.CaseID, caseID)
	}

	if profile.CommitID != commitID {
		t.Errorf("NewSuspectProfile() CommitID = %v, want %v", profile.CommitID, commitID)
	}

	if profile.Attributes == nil {
		t.Error("NewSuspectProfile() should initialize Attributes")
	}

	if profile.UpdatedAt.IsZero() {
		t.Error("NewSuspectProfile() should set UpdatedAt")
	}

	if err := profile.Validate(); err != nil {
		t.Errorf("NewSuspectProfile() created invalid profile: %v", err)
	}
}

func TestRangeAttribute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		attr    *RangeAttribute
		wantErr bool
	}{
		{
			name: "valid range",
			attr: &RangeAttribute{
				Min:        25,
				Max:        35,
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "min equals max (valid)",
			attr: &RangeAttribute{
				Min:        30,
				Max:        30,
				Confidence: 0.9,
			},
			wantErr: false,
		},
		{
			name: "min greater than max should fail",
			attr: &RangeAttribute{
				Min:        35,
				Max:        25,
				Confidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "confidence below 0 should fail",
			attr: &RangeAttribute{
				Min:        25,
				Max:        35,
				Confidence: -0.1,
			},
			wantErr: true,
		},
		{
			name: "confidence above 1 should fail",
			attr: &RangeAttribute{
				Min:        25,
				Max:        35,
				Confidence: 1.1,
			},
			wantErr: true,
		},
		{
			name: "confidence at boundaries (0 and 1)",
			attr: &RangeAttribute{
				Min:        25,
				Max:        35,
				Confidence: 1.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RangeAttribute.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringAttribute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		attr    *StringAttribute
		wantErr bool
	}{
		{
			name: "valid attribute",
			attr: &StringAttribute{
				Value:      "average",
				Confidence: 0.7,
			},
			wantErr: false,
		},
		{
			name: "confidence below 0 should fail",
			attr: &StringAttribute{
				Value:      "slim",
				Confidence: -0.5,
			},
			wantErr: true,
		},
		{
			name: "confidence above 1 should fail",
			attr: &StringAttribute{
				Value:      "heavy",
				Confidence: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("StringAttribute.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuspectAttributes_Validate(t *testing.T) {
	tests := []struct {
		name    string
		attrs   *SuspectAttributes
		wantErr bool
	}{
		{
			name:    "empty attributes (valid)",
			attrs:   NewEmptySuspectAttributes(),
			wantErr: false,
		},
		{
			name: "invalid age_range should fail",
			attrs: &SuspectAttributes{
				AgeRange: &RangeAttribute{
					Min:        50,
					Max:        20, // Invalid: min > max
					Confidence: 0.8,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid height_range should fail",
			attrs: &SuspectAttributes{
				HeightRangeCm: &RangeAttribute{
					Min:        180,
					Max:        170, // Invalid: min > max
					Confidence: 0.8,
				},
			},
			wantErr: true,
		},
		{
			name: "valid complete attributes",
			attrs: &SuspectAttributes{
				AgeRange: &RangeAttribute{
					Min:        25,
					Max:        35,
					Confidence: 0.7,
				},
				HeightRangeCm: &RangeAttribute{
					Min:        170,
					Max:        180,
					Confidence: 0.8,
				},
				Build: &StringAttribute{
					Value:      "average",
					Confidence: 0.6,
				},
				Hair: &HairAttribute{
					Style:      "short",
					Color:      "black",
					Confidence: 0.75,
				},
				DistinctiveFeatures: []FeatureAttribute{
					{
						Description: "Scar on left cheek",
						Confidence:  0.9,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attrs.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SuspectAttributes.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWitnessStatementInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		stmt    *WitnessStatementInput
		wantErr bool
	}{
		{
			name:    "empty statement should fail",
			stmt:    &WitnessStatementInput{},
			wantErr: true,
		},
		{
			name: "missing source_name should fail",
			stmt: &WitnessStatementInput{
				Content:     "Suspect was tall",
				Credibility: 0.8,
			},
			wantErr: true,
		},
		{
			name: "missing content should fail",
			stmt: &WitnessStatementInput{
				SourceName:  "Witness A",
				Credibility: 0.8,
			},
			wantErr: true,
		},
		{
			name: "credibility below 0 should fail",
			stmt: &WitnessStatementInput{
				SourceName:  "Witness A",
				Content:     "Suspect was tall",
				Credibility: -0.1,
			},
			wantErr: true,
		},
		{
			name: "credibility above 1 should fail",
			stmt: &WitnessStatementInput{
				SourceName:  "Witness A",
				Content:     "Suspect was tall",
				Credibility: 1.5,
			},
			wantErr: true,
		},
		{
			name: "valid statement",
			stmt: &WitnessStatementInput{
				SourceName:  "Witness A",
				Content:     "Suspect was approximately 175cm, short hair, wearing glasses",
				Credibility: 0.8,
			},
			wantErr: false,
		},
		{
			name: "valid statement with 0 credibility",
			stmt: &WitnessStatementInput{
				SourceName:  "Anonymous Tip",
				Content:     "Some description",
				Credibility: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stmt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WitnessStatementInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEmptySuspectAttributes(t *testing.T) {
	attrs := NewEmptySuspectAttributes()

	if attrs == nil {
		t.Fatal("NewEmptySuspectAttributes() should not return nil")
	}

	if attrs.DistinctiveFeatures == nil {
		t.Error("NewEmptySuspectAttributes() should initialize DistinctiveFeatures slice")
	}

	if len(attrs.DistinctiveFeatures) != 0 {
		t.Error("NewEmptySuspectAttributes() DistinctiveFeatures should be empty")
	}

	// All optional fields should be nil
	if attrs.AgeRange != nil {
		t.Error("NewEmptySuspectAttributes() AgeRange should be nil")
	}
	if attrs.HeightRangeCm != nil {
		t.Error("NewEmptySuspectAttributes() HeightRangeCm should be nil")
	}
	if attrs.Build != nil {
		t.Error("NewEmptySuspectAttributes() Build should be nil")
	}
}
