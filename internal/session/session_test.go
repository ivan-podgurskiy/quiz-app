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
