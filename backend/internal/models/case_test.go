package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestCase_Validate(t *testing.T) {
	tests := []struct {
		name    string
		c       *Case
		wantErr bool
	}{
		{
			name:    "empty case should fail",
			c:       &Case{},
			wantErr: true,
		},
		{
			name: "missing title should fail",
			c: &Case{
				ID:          uuid.New(),
				Description: "Some description",
			},
			wantErr: true,
		},
		{
			name: "title too long should fail",
			c: &Case{
				ID:    uuid.New(),
				Title: string(make([]byte, 201)), // 201 characters
			},
			wantErr: true,
		},
		{
			name: "valid case with title only",
			c: &Case{
				ID:    uuid.New(),
				Title: "Test Case",
			},
			wantErr: false,
		},
		{
			name: "valid case with all fields",
			c: &Case{
				ID:          uuid.New(),
				Title:       "Case A-2025-001",
				Description: "Residential burglary investigation",
			},
			wantErr: false,
		},
		{
			name: "title at max length (200 chars)",
			c: &Case{
				ID:    uuid.New(),
				Title: string(make([]byte, 200)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Case.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCase(t *testing.T) {
	title := "Test Case"
	description := "Test Description"

	c := NewCase(title, description)

	if c.ID == uuid.Nil {
		t.Error("NewCase() should generate a non-nil UUID")
	}

	if c.Title != title {
		t.Errorf("NewCase() Title = %v, want %v", c.Title, title)
	}

	if c.Description != description {
		t.Errorf("NewCase() Description = %v, want %v", c.Description, description)
	}

	if c.CreatedAt.IsZero() {
		t.Error("NewCase() should set CreatedAt")
	}

	// Validate the created case
	if err := c.Validate(); err != nil {
		t.Errorf("NewCase() created invalid case: %v", err)
	}
}

func TestNewCase_UniqueIDs(t *testing.T) {
	c1 := NewCase("Case 1", "")
	c2 := NewCase("Case 2", "")

	if c1.ID == c2.ID {
		t.Error("NewCase() should generate unique IDs")
	}
}
