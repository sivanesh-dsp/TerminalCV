package resume

import "strings"

// SearchHit is a single match produced by Search, tagged with the portfolio
// section it came from so the UI can group results.
type SearchHit struct {
	Section string // uppercase section label, e.g. "EXPERIENCE"
	Title   string // the matched item's headline
	Snippet string // surrounding context for the match
}

// Search performs a case-insensitive substring search across every résumé
// section and returns grouped hits in a stable section order. It is pure
// business logic shared by any renderer (TUI today, reusable elsewhere).
func (r *Resume) Search(query string) []SearchHit {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return nil
	}
	var hits []SearchHit
	add := func(section, title, snippet string) {
		hits = append(hits, SearchHit{Section: section, Title: title, Snippet: snippet})
	}
	match := func(s string) bool { return strings.Contains(strings.ToLower(s), q) }

	// ABOUT
	if match(r.Summary) || match(r.Title) || match(r.Name) {
		add("ABOUT", r.Name, r.Title)
	}

	// EXPERIENCE
	for _, e := range r.Experience {
		if match(e.Role) || match(e.Company) || match(e.Location) {
			add("EXPERIENCE", e.Role+" — "+e.Company, e.Start+" – "+e.End)
			continue
		}
		for _, h := range e.Highlights {
			if match(h) {
				add("EXPERIENCE", e.Role+" — "+e.Company, snippet(h, q))
				break
			}
		}
	}

	// PROJECTS
	for _, p := range r.Projects {
		if match(p.Name) || match(p.Description) || anyMatch(p.Tech, q) {
			add("PROJECTS", p.Name, snippet(p.Description, q))
		}
	}

	// SKILLS
	for _, c := range r.Skills {
		for _, s := range c.Skills {
			if match(s) || match(c.Name) {
				add("SKILLS", c.Name, s)
			}
		}
	}

	// CERTIFICATIONS
	for _, c := range r.Certifications {
		if match(c.Name) || match(c.Issuer) {
			add("CERTIFICATIONS", c.Name, c.Issuer)
		}
	}

	// EDUCATION
	for _, ed := range r.Education {
		if match(ed.Degree) || match(ed.Institution) || match(ed.Location) {
			add("EDUCATION", ed.Degree, ed.Institution)
		}
	}

	// ACHIEVEMENTS
	for _, a := range r.Achievements {
		if match(a) {
			add("ACHIEVEMENTS", snippet(a, q), "")
		}
	}

	// TIMELINE
	for _, t := range r.Timeline {
		if match(t.Title) || match(t.Subtitle) || match(t.Date) {
			add("TIMELINE", t.Date+" — "+t.Title, t.Subtitle)
		}
	}

	// CONTACT
	if match(r.Contact.Email) || match(r.Contact.Location) {
		add("CONTACT", r.Contact.Email, r.Contact.Location)
	}

	return hits
}

func anyMatch(items []string, q string) bool {
	for _, s := range items {
		if strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}

// snippet returns a trimmed window of text around the first match of q so long
// paragraphs don't overflow the results list.
func snippet(text, q string) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, q)
	if idx < 0 {
		if len(text) > 80 {
			return strings.TrimSpace(text[:80]) + "…"
		}
		return text
	}
	const pad = 32
	start := idx - pad
	if start < 0 {
		start = 0
	}
	end := idx + len(q) + pad
	if end > len(text) {
		end = len(text)
	}
	out := strings.TrimSpace(text[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(text) {
		out += "…"
	}
	return out
}
