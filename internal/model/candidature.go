package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"uuid"
)

type Status string

const (
	StatusApplied   Status = "applied"
	StatusInterview Status = "interview"
	StatusOffer     Status = "offer"
	StatusRejected  Status = "rejected"
)

// Valid indique si le statut fait partie des valeurs autorisées.
func (s Status) Valid() bool {
	switch s {
	case StatusApplied, StatusInterview, StatusOffer, StatusRejected:
		return true
	default:
		return false
	}
}

type Candidature struct {
	ID        uuid.UUID `json:"id"`
	Company   string    `json:"company"`
	Position  string    `json:"position"`
	URL       string    `json:"url,omitempty"`
	AppliedAt time.Time `json:"applied_at"`
	Status    Status    `json:"status"`
	Notes     string    `json:"notes,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// New construit une candidature valide avec un UUID v7 et les timestamps courants.
func New(company, position, link, notes string) (Candidature, error) {
	id := uuid.NewV7()

	now := time.Now()
	c := Candidature{
		ID:        id,
		Company:   strings.TrimSpace(company),
		Position:  strings.TrimSpace(position),
		URL:       strings.TrimSpace(link),
		AppliedAt: now,
		Status:    StatusApplied,
		Notes:     strings.TrimSpace(notes),
		UpdatedAt: now,
	}

	if err := c.Validate(); err != nil {
		return Candidature{}, err
	}
	return c, nil
}

// Validate vérifie les invariants d'une candidature.
func (c Candidature) Validate() error {
	if c.Company == "" {
		return fmt.Errorf("company is required")
	}
	if c.Position == "" {
		return fmt.Errorf("position is required")
	}
	if !c.Status.Valid() {
		return fmt.Errorf("invalid status %q", c.Status)
	}
	if c.URL != "" {
		u, err := url.Parse(c.URL)
		if err != nil {
			return fmt.Errorf("invalid url: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("url must use http or https")
		}
		if u.Host == "" {
			return fmt.Errorf("url must have a host")
		}
	}
	return nil
}
