package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Record holds the SM-2 state for a single question.
type Record struct {
	LastSeen     string  `json:"last_seen"`
	NextDue      string  `json:"next_due"`
	TimesSeen    int     `json:"times_seen"`
	TimesCorrect int     `json:"times_correct"`
	EaseFactor   float64 `json:"ease_factor"`
	IntervalDays float64 `json:"interval_days"`
}

// Store is a map from question ID to its progress Record.
type Store map[string]*Record

// progressPath returns ~/.quiz/progress.json
func progressPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, ".quiz", "progress.json"), nil
}

// Load reads the progress file. Returns an empty Store if the file doesn't exist.
func Load() (Store, error) {
	path, err := progressPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(Store), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading progress file: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parsing progress file: %w", err)
	}
	if store == nil {
		store = make(Store)
	}
	return store, nil
}

// Save atomically writes the store to ~/.quiz/progress.json.
func (s Store) Save() error {
	path, err := progressPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating .quiz dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling progress: %w", err)
	}

	// Write to temp file then rename for atomicity
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing temp progress file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming progress file: %w", err)
	}
	return nil
}

// Get returns the Record for a question ID, or nil if not found.
func (s Store) Get(id string) *Record {
	return s[id]
}

// marshalStore serializes the store to JSON (used by Save and tests).
func marshalStore(s Store) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// loadFromPath reads and parses a progress JSON file at a specific path (used by tests).
func loadFromPath(path string) (Store, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(Store), nil
	}
	if err != nil {
		return nil, err
	}
	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store == nil {
		store = make(Store)
	}
	return store, nil
}
