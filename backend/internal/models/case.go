package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Case represents a crime investigation case
type Case struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Validate checks if the Case is valid
func (c *Case) Validate() error {
	if c.Title == "" {
		return errors.New("title is required")
	}
	if len(c.Title) > 200 {
		return errors.New("title must be 200 characters or less")
	}
	return nil
}

// NewCase creates a new Case with generated ID and timestamp
func NewCase(title, description string) *Case {
	return &Case{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}
