// Package resume defines the portfolio data model and loads it at runtime from
// the shared, single-source-of-truth file (shared/resume.json). The same file
// is consumed by the React website, so editing it updates both frontends.
package resume

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type SocialLink struct {
	Handle string `json:"handle"`
	URL    string `json:"url"`
}

type Contact struct {
	Email     string      `json:"email"`
	Phone     string      `json:"phone,omitempty"`
	Location  string      `json:"location"`
	LinkedIn  *SocialLink `json:"linkedin,omitempty"`
	GitHub    *SocialLink `json:"github,omitempty"`
	Portfolio *SocialLink `json:"portfolio,omitempty"`
}

type SkillCategory struct {
	Name   string   `json:"name"`
	Skills []string `json:"skills"`
}

type Experience struct {
	Role       string   `json:"role"`
	Company    string   `json:"company"`
	Start      string   `json:"start"`
	End        string   `json:"end"`
	Location   string   `json:"location"`
	Highlights []string `json:"highlights"`
}

type Certification struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer,omitempty"`
}

type Education struct {
	Degree      string `json:"degree"`
	Institution string `json:"institution"`
	Location    string `json:"location,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
}

type TimelineEvent struct {
	Date     string `json:"date"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
}

// Resume mirrors shared/resume.json field-for-field.
type Resume struct {
	Name           string          `json:"name"`
	Title          string          `json:"title"`
	Username       string          `json:"username"`
	Host           string          `json:"host"`
	Summary        string          `json:"summary"`
	Contact        Contact         `json:"contact"`
	Skills         []SkillCategory `json:"skills"`
	Experience     []Experience    `json:"experience"`
	Certifications []Certification `json:"certifications"`
	Education      []Education     `json:"education"`
	Projects       []Project       `json:"projects"`
	Achievements   []string        `json:"achievements"`
	Timeline       []TimelineEvent `json:"timeline"`
	CareerStartISO string          `json:"careerStartISO"`
	ResumeFile     string          `json:"resumeFile"`
}

// candidatePaths lists where the shared résumé JSON may live, in priority order.
func candidatePaths(explicit string) []string {
	paths := []string{}
	if explicit != "" {
		paths = append(paths, explicit)
	}
	if env := os.Getenv("RESUME_PATH"); env != "" {
		paths = append(paths, env)
	}
	return append(paths,
		"shared/resume.json",
		"../shared/resume.json",
		"../../shared/resume.json",
		"../../../shared/resume.json",
		"/app/shared/resume.json",
	)
}

// Load reads and parses the résumé from the first path that exists.
func Load(explicit string) (*Resume, error) {
	tried := candidatePaths(explicit)
	for _, p := range tried {
		if p == "" {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		return Parse(b)
	}
	return nil, fmt.Errorf("resume.json not found (looked in %v); set RESUME_PATH", tried)
}

// Parse decodes résumé JSON and validates the essentials. Unknown fields are
// tolerated so the shared file can grow web-only sections without breaking SSH.
func Parse(b []byte) (*Resume, error) {
	var r Resume
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parse resume.json: %w", err)
	}
	if r.Name == "" || len(r.Skills) == 0 || len(r.Experience) == 0 {
		return nil, fmt.Errorf("resume.json is missing required fields (name/skills/experience)")
	}
	return &r, nil
}

// AllTechnologies returns the de-duplicated union of every skill.
func (r *Resume) AllTechnologies() []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range r.Skills {
		for _, s := range c.Skills {
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	return out
}

// ExperienceMonths returns whole months of experience since CareerStartISO.
func (r *Resume) ExperienceMonths(now time.Time) int {
	start, err := time.Parse("2006-01-02", r.CareerStartISO)
	if err != nil {
		return 0
	}
	months := (now.Year()-start.Year())*12 + int(now.Month()) - int(start.Month())
	if now.Day() < start.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}
	return months
}

// ExperienceLabel renders a human duration like "2 yrs 1 mo".
func (r *Resume) ExperienceLabel(now time.Time) string {
	m := r.ExperienceMonths(now)
	years, months := m/12, m%12
	label := ""
	if years > 0 {
		label += fmt.Sprintf("%d yr", years)
		if years > 1 {
			label += "s"
		}
		label += " "
	}
	label += fmt.Sprintf("%d mo", months)
	if months != 1 {
		label += "s"
	}
	return label
}

// Employers counts distinct companies in the experience list.
func (r *Resume) Employers() int {
	seen := map[string]bool{}
	for _, e := range r.Experience {
		seen[e.Company] = true
	}
	return len(seen)
}
