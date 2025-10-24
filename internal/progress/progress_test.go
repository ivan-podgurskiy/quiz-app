package progress

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecord_Update_Correct(t *testing.T) {
	r := &Record{
		EaseFactor:   2.5,
		IntervalDays: 1.0,
	}
	today := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	r.Update(true, today)

	if r.TimesSeen != 1 {
		t.Errorf("expected TimesSeen=1, got %d", r.TimesSeen)
	}
	if r.TimesCorrect != 1 {
		t.Errorf("expected TimesCorrect=1, got %d", r.TimesCorrect)
	}
	// interval = max(1, 1.0 * 2.5) = 2.5, but we floor to int when adding days
	if r.IntervalDays != 2.5 {
		t.Errorf("expected IntervalDays=2.5, got %f", r.IntervalDays)
	}
	if r.EaseFactor != 2.5 { // min(2.5, 2.5+0.1) = 2.5
		t.Errorf("expected EaseFactor=2.5 (capped), got %f", r.EaseFactor)
	}
	if r.LastSeen != "2026-03-02" {
		t.Errorf("unexpected LastSeen: %s", r.LastSeen)
	}
}

func TestRecord_Update_Wrong(t *testing.T) {
	r := &Record{
		EaseFactor:   2.5,
		IntervalDays: 10.0,
	}
	today := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	r.Update(false, today)

	if r.TimesCorrect != 0 {
		t.Errorf("expected TimesCorrect=0, got %d", r.TimesCorrect)
	}
	if r.IntervalDays != 1.0 {
		t.Errorf("expected IntervalDays reset to 1.0, got %f", r.IntervalDays)
	}
	if r.EaseFactor != 2.3 {
		t.Errorf("expected EaseFactor=2.3, got %f", r.EaseFactor)
	}
}

func TestRecord_Update_EaseFactorFloor(t *testing.T) {
	r := &Record{
		EaseFactor:   1.4,
		IntervalDays: 1.0,
	}
	today := time.Now()
	r.Update(false, today)
	if r.EaseFactor < minEaseFactor {
		t.Errorf("EaseFactor %f dropped below floor %f", r.EaseFactor, minEaseFactor)
	}
}

func TestRecord_Update_InitializesDefaults(t *testing.T) {
	r := &Record{} // zero value
	today := time.Now()
	r.Update(true, today)
	if r.EaseFactor == 0 {
		t.Error("EaseFactor should be initialized")
	}
	if r.IntervalDays == 0 {
		t.Error("IntervalDays should be initialized")
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	// Override home dir by writing directly to a temp path
	tmp := t.TempDir()
	path := filepath.Join(tmp, "progress.json")

	store := make(Store)
	store["q-001"] = &Record{
		TimesSeen:    3,
		TimesCorrect: 2,
		EaseFactor:   2.5,
		IntervalDays: 4.0,
		LastSeen:     "2026-03-01",
		NextDue:      "2026-03-05",
	}

	// Write directly to temp path for testing
	data, err := marshalStore(store)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Read back
	loaded, err := loadFromPath(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	rec := loaded.Get("q-001")
	if rec == nil {
		t.Fatal("expected record for q-001")
	}
	if rec.TimesSeen != 3 {
		t.Errorf("expected TimesSeen=3, got %d", rec.TimesSeen)
	}
	if rec.EaseFactor != 2.5 {
		t.Errorf("expected EaseFactor=2.5, got %f", rec.EaseFactor)
	}
}

func TestStore_Get_Missing(t *testing.T) {
	store := make(Store)
	if store.Get("nonexistent") != nil {
		t.Error("expected nil for missing key")
	}
}
