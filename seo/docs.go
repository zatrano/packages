package seo

import (
	"os"
	"time"

	"github.com/zatrano/packages/docs"
)

// AttachDocs adds documentation pages to the sitemap and llms-full inventory.
// Safe to call when repo is nil. Duplicate locations are merged.
func (s *Site) AttachDocs(repo *docs.Repository) {
	if s == nil || repo == nil {
		return
	}
	s.mu.Lock()
	s.docs = repo
	prefix := s.opts.DocsPrefix
	s.mu.Unlock()

	pages, err := repo.List()
	if err != nil {
		return
	}
	now := time.Now().UTC()
	for _, page := range pages {
		loc := prefix
		if page.Slug != "" && page.Slug != "index" {
			loc = prefix + "/" + page.Slug
		}
		priority := 0.6
		freq := "weekly"
		lastMod := now
		if page.Path != "" {
			if info, statErr := os.Stat(page.Path); statErr == nil {
				lastMod = info.ModTime().UTC()
			}
		}
		s.Add(loc, URL{Priority: priority, ChangeFreq: freq, LastMod: lastMod})
	}
}
