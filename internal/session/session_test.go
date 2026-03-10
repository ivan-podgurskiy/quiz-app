package session

import (
	"testing"

	"quiz-app/internal/question"
)

func makeQ(id string) question.Question {
	return question.Question{
		ID:   id,
		Type: question.SingleChoice,
		Options: []question.Option{
			{ID: "a", Text: "A", Correct: true},
			{ID: "b", Text: "B", Correct: false},
		},
	}
}

func TestStore_CreateAndGet(t *testing.T) {
	store := NewStore()
	questions := []question.Question{makeQ("q1"), makeQ("q2"), makeQ("q3")}

	sess := store.Create(questions)
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
	if sess.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if len(sess.Questions) != 3 {
		t.Errorf("expected 3 questions, got %d", len(sess.Questions))
	}
	if sess.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex=0, got %d", sess.CurrentIndex)
	}

	got := store.Get(sess.ID)
	if got == nil {
		t.Fatal("expected to retrieve session by ID")
	}
	if got.ID != sess.ID {
		t.Errorf("retrieved wrong session: %s != %s", got.ID, sess.ID)
	}
}

func TestStore_Get_Missing(t *testing.T) {
	store := NewStore()
	if store.Get("nonexistent") != nil {
		t.Error("expected nil for missing session")
	}
}

func TestStore_Delete(t *testing.T) {
	store := NewStore()
	sess := store.Create([]question.Question{makeQ("q1")})
	store.Delete(sess.ID)
	if store.Get(sess.ID) != nil {
		t.Error("expected nil after deletion")
	}
}

func TestStore_ShuffledOpts(t *testing.T) {
	store := NewStore()
	questions := []question.Question{makeQ("q1"), makeQ("q2")}
	sess := store.Create(questions)

	for _, q := range questions {
		opts, ok := sess.ShuffledOpts[q.ID]
		if !ok {
			t.Errorf("missing shuffled opts for question %s", q.ID)
		}
		if len(opts) != 2 {
			t.Errorf("expected 2 options for %s, got %d", q.ID, len(opts))
		}
	}
}

// optionOrderSignature returns the display order of option IDs so we can detect
// when the same question shows options in a different order (different visual position).
func optionOrderSignature(opts []question.Option) string {
	s := make([]byte, 0, len(opts)*2)
	for _, o := range opts {
		s = append(s, o.ID...)
	}
	return string(s)
}

// TestStore_OptionOrderRotatesAcrossSessions verifies that the order of answer options
// (and thus the visual position of the correct answer) varies between sessions, so users
// cannot rely on remembering "the answer was always the second option".
func TestStore_OptionOrderRotatesAcrossSessions(t *testing.T) {
	// Question with 4 options so shuffle can produce many orderings.
	q := question.Question{
		ID:   "q1",
		Type: question.SingleChoice,
		Options: []question.Option{
			{ID: "a", Text: "A", Correct: true},
			{ID: "b", Text: "B", Correct: false},
			{ID: "c", Text: "C", Correct: false},
			{ID: "d", Text: "D", Correct: false},
		},
	}
	store := NewStore()
	seen := make(map[string]bool)
	for i := 0; i < 30; i++ {
		sess := store.Create([]question.Question{q})
		opts := sess.ShuffledOpts[q.ID]
		if len(opts) != 4 {
			t.Fatalf("session %d: expected 4 options, got %d", i, len(opts))
		}
		seen[optionOrderSignature(opts)] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected at least 2 different option orderings across 30 sessions (correct answer should not always be in the same place), got %d",
			len(seen))
	}
}

// TestStore_DisplayOrderPreserved verifies that the order of questions in the session
// is the display order shown to the user, and that different runs can have different display orders.
func TestStore_DisplayOrderPreserved(t *testing.T) {
	store := NewStore()
	q1, q2, q3 := makeQ("q1"), makeQ("q2"), makeQ("q3")

	// Simulate two quiz runs with different orderings (e.g. from SelectQuestions).
	orderA := []question.Question{q1, q2, q3}
	orderB := []question.Question{q2, q1, q3}

	sessA := store.Create(orderA)
	sessB := store.Create(orderB)

	// Display order is sess.Questions[CurrentIndex]; first question shown must match input order.
	if sessA.Questions[0].ID != "q1" || sessB.Questions[0].ID != "q2" {
		t.Errorf("display order not preserved: sessA first=%s (want q1), sessB first=%s (want q2)",
			sessA.Questions[0].ID, sessB.Questions[0].ID)
	}
}

func TestStore_UniqueIDs(t *testing.T) {
	store := NewStore()
	ids := map[string]bool{}
	for i := 0; i < 20; i++ {
		sess := store.Create([]question.Question{makeQ("q1")})
		if ids[sess.ID] {
			t.Errorf("duplicate session ID generated: %s", sess.ID)
		}
		ids[sess.ID] = true
	}
}

func TestNewUUID_Length(t *testing.T) {
	id := newUUID()
	if len(id) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected UUID length 32, got %d", len(id))
	}
}
