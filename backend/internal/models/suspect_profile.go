package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SuspectProfile represents a suspect's structured attributes
type SuspectProfile struct {
	CaseID           uuid.UUID          `json:"case_id"`
	CommitID         uuid.UUID          `json:"commit_id"`
	Attributes       *SuspectAttributes `json:"attributes"`
	PortraitAssetKey string             `json:"portrait_asset_key,omitempty"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// Validate checks if the SuspectProfile is valid
func (sp *SuspectProfile) Validate() error {
	if sp.CaseID == uuid.Nil {
		return errors.New("case_id is required")
	}
	if sp.CommitID == uuid.Nil {
		return errors.New("commit_id is required")
	}
	if sp.Attributes != nil {
		if err := sp.Attributes.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// NewSuspectProfile creates a new SuspectProfile
func NewSuspectProfile(caseID, commitID uuid.UUID) *SuspectProfile {
	return &SuspectProfile{
		CaseID:     caseID,
		CommitID:   commitID,
		Attributes: NewEmptySuspectAttributes(),
		UpdatedAt:  time.Now().UTC(),
	}
}

// SuspectAttributes represents structured physical attributes
type SuspectAttributes struct {
	AgeRange            *RangeAttribute     `json:"age_range,omitempty"`
	HeightRangeCm       *RangeAttribute     `json:"height_range_cm,omitempty"`
	Build               *StringAttribute    `json:"build,omitempty"`
	SkinTone            *StringAttribute    `json:"skin_tone,omitempty"`
	Hair                *HairAttribute      `json:"hair,omitempty"`
	FacialHair          *StringAttribute    `json:"facial_hair,omitempty"`
	Glasses             *StringAttribute    `json:"glasses,omitempty"`
	DistinctiveFeatures []FeatureAttribute  `json:"distinctive_features,omitempty"`
}

// Validate checks if the SuspectAttributes are valid
func (sa *SuspectAttributes) Validate() error {
	if sa.AgeRange != nil {
		if err := sa.AgeRange.Validate(); err != nil {
			return errors.New("age_range: " + err.Error())
		}
	}
	if sa.HeightRangeCm != nil {
		if err := sa.HeightRangeCm.Validate(); err != nil {
			return errors.New("height_range_cm: " + err.Error())
		}
	}
	return nil
}

// NewEmptySuspectAttributes creates empty suspect attributes
func NewEmptySuspectAttributes() *SuspectAttributes {
	return &SuspectAttributes{
		DistinctiveFeatures: []FeatureAttribute{},
	}
}

// RangeAttribute represents a numeric range with confidence
type RangeAttribute struct {
	Min              float64           `json:"min"`
	Max              float64           `json:"max"`
	Confidence       float64           `json:"confidence"`
	SupportingSources []AttributeSource `json:"supporting_sources,omitempty"`
	ConflictSources   []AttributeSource `json:"conflict_sources,omitempty"`
}

// Validate checks if the RangeAttribute is valid
func (ra *RangeAttribute) Validate() error {
	if ra.Min > ra.Max {
		return errors.New("min must be less than or equal to max")
	}
	if ra.Confidence < 0 || ra.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	return nil
}

// StringAttribute represents a string value with confidence
type StringAttribute struct {
	Value             string            `json:"value"`
	Confidence        float64           `json:"confidence"`
	SupportingSources []AttributeSource `json:"supporting_sources,omitempty"`
	ConflictSources   []AttributeSource `json:"conflict_sources,omitempty"`
}

// Validate checks if the StringAttribute is valid
func (sa *StringAttribute) Validate() error {
	if sa.Confidence < 0 || sa.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}
	return nil
}

// HairAttribute represents hair description
type HairAttribute struct {
	Style             string            `json:"style"`
	Color             string            `json:"color"`
	Confidence        float64           `json:"confidence"`
	SupportingSources []AttributeSource `json:"supporting_sources,omitempty"`
	ConflictSources   []AttributeSource `json:"conflict_sources,omitempty"`
}

// FeatureAttribute represents a distinctive feature
type FeatureAttribute struct {
	Description       string            `json:"description"`
	Confidence        float64           `json:"confidence"`
	SupportingSources []AttributeSource `json:"supporting_sources,omitempty"`
}

// AttributeSource represents the source of an attribute
type AttributeSource struct {
	CommitID    string  `json:"commit_id"`
	SourceName  string  `json:"source_name"`
	Credibility float64 `json:"credibility"`
}

// WitnessStatementInput represents input from witness statements
type WitnessStatementInput struct {
	SourceName  string  `json:"source_name"`
	Content     string  `json:"content"`
	Credibility float64 `json:"credibility"`
}

// Validate checks if the WitnessStatementInput is valid
func (ws *WitnessStatementInput) Validate() error {
	if ws.SourceName == "" {
		return errors.New("source_name is required")
	}
	if ws.Content == "" {
		return errors.New("content is required")
	}
	if ws.Credibility < 0 || ws.Credibility > 1 {
		return errors.New("credibility must be between 0 and 1")
	}
	return nil
}
