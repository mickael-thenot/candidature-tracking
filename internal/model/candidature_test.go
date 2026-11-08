package model

import (
	"strings"
	"testing"
)

func TestNew_Valid(t *testing.T) {
	c, err := New(" Google ", " Dev ", "https://google.com", " Note ")
	if err != nil {
		t.Fatalf("New() returned an unexpected error: %v", err)
	}

	// Vérification du nettoyage des espaces (TrimSpace)
	if c.Company != "Google" {
		t.Errorf("expected company 'Google', got %q", c.Company)
	}
	if c.Position != "Dev" {
		t.Errorf("expected position 'Dev', got %q", c.Position)
	}
	if c.Notes != "Note" {
		t.Errorf("expected notes 'Note', got %q", c.Notes)
	}

	// Vérification du statut par défaut
	if c.Status != StatusApplied {
		t.Errorf("expected default status %q, got %q", StatusApplied, c.Status)
	}

	// On vérifie que le timestamp n'est pas nul
	if c.AppliedAt.IsZero() {
		t.Error("expected AppliedAt to be set, got zero time")
	}
}

func TestNew_InvalidCompany(t *testing.T) {
	// On teste avec une chaîne vide et une chaîne d'espaces
	testCases := []string{"", "  "}

	for _, tc := range testCases {
		_, err := New(tc, "Dev", "", "")
		if err == nil {
			t.Errorf("expected error for company %q, got nil", tc)
			continue
		}
		if !strings.Contains(err.Error(), "company is required") {
			t.Errorf("expected 'company is required' error, got %q", err.Error())
		}
	}
}

func TestNew_InvalidURL(t *testing.T) {
	// Le modèle exige http ou https
	invalidURLs := []string{
		"pas-une-url",
		"ftp://google.com",
		"https://",
	}

	for _, u := range invalidURLs {
		_, err := New("Google", "Dev", u, "")
		if err == nil {
			t.Errorf("expected error for URL %q, got nil", u)
		}
	}
}

func TestValidate_InvalidStatus(t *testing.T) {
	// On crée une candidature manuellement pour contourner le constructeur New
	c := Candidature{
		Company:  "Google",
		Position: "Dev",
		Status:   Status("statut-inconnu"),
	}

	err := c.Validate()
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' error, got %q", err.Error())
	}
}
