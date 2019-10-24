package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
)

type recentEntry struct {
	Author string `json:"author"`
	Date   string `json:"date"`
	// Skip JSON serialization for lastModified as clients don't need this field.
	lastModified string
	Markdown     string `json:"markdown"`
}

func (s *defaultServer) recentEntriesGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.datastore.Users()
		if err != nil {
			log.Printf("Failed to retrieve users: %s", err)
			http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
			return
		}

		entries := []recentEntry{}
		for _, username := range users {
			userEntries, err := s.datastore.GetEntries(username)
			if err != nil {
				log.Printf("Failed to retrieve entries for user %s: %s", username, err)
				http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
				return
			}
			for _, entry := range userEntries {
				// Filter low-effort posts or test posts from the recent list.
				const minimumRelevantLength = 30
				if len(entry.Markdown) < minimumRelevantLength {
					continue
				}
				entries = append(entries, recentEntry{
					Author:       username,
					Date:         entry.Date,
					lastModified: entry.LastModified,
					Markdown:     entry.Markdown,
				})
			}
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Date < entries[j].Date {
				return true
			}
			if entries[i].Date > entries[j].Date {
				return false
			}
			return entries[i].lastModified < entries[j].lastModified
		})

		// Reverse the order of entries.
		for i := len(entries)/2 - 1; i >= 0; i-- {
			opp := len(entries) - 1 - i
			entries[i], entries[opp] = entries[opp], entries[i]
		}

		const maxEntries = 15
		if len(entries) > maxEntries {
			entries = entries[:maxEntries]
		}

		if err := json.NewEncoder(w).Encode(entries); err != nil {
			panic(err)
		}
	}
}
