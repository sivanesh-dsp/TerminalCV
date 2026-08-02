package resume

import (
	"testing"
	"time"
)

const sharedPath = "../../../shared/resume.json"

func TestLoadSharedResume(t *testing.T) {
	r, err := Load(sharedPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.Name != "Sivanesh B" {
		t.Errorf("Name = %q", r.Name)
	}
	if r.Username == "" || r.Host == "" {
		t.Errorf("username/host empty")
	}
	if len(r.Skills) == 0 || len(r.Experience) == 0 || len(r.Certifications) == 0 {
		t.Errorf("missing sections: skills=%d exp=%d certs=%d",
			len(r.Skills), len(r.Experience), len(r.Certifications))
	}
	if r.Contact.Email == "" {
		t.Errorf("email empty")
	}
	// GitHub is intentionally absent on the résumé.
	if r.Contact.GitHub != nil {
		t.Errorf("GitHub should be nil (not on résumé)")
	}
}

func TestAllTechnologiesDeduped(t *testing.T) {
	r, err := Load(sharedPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	techs := r.AllTechnologies()
	if len(techs) < 20 {
		t.Errorf("expected many technologies, got %d", len(techs))
	}
	seen := map[string]bool{}
	for _, tch := range techs {
		if seen[tch] {
			t.Errorf("duplicate technology: %s", tch)
		}
		seen[tch] = true
	}
}

func TestExperienceLabel(t *testing.T) {
	r, err := Load(sharedPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// From 2024-08-01 to 2026-08-02 is exactly 2 years.
	got := r.ExperienceLabel(time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC))
	if got != "2 yrs 0 mos" {
		t.Errorf("ExperienceLabel = %q, want %q", got, "2 yrs 0 mos")
	}
	if r.Employers() != 1 {
		t.Errorf("Employers = %d, want 1", r.Employers())
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	if _, err := Parse([]byte(`{"name":""}`)); err == nil {
		t.Errorf("expected error for empty résumé")
	}
	if _, err := Parse([]byte(`not json`)); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}
