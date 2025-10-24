package progress

import (
	"testing"
	"time"

	"quiz-app/internal/question"
)

func makeQuestion(id, topic, difficulty string) question.Question {
	return question.Question{
		ID:         id,
		Topic:      topic,
		Difficulty: difficulty,
		Type:       question.SingleChoice,
		Question:   "Q?",
		Options: []question.Option{
			{ID: "a", Text: "A", Correct: true},
			{ID: "b", Text: "B", Correct: false},
		},
	}
}

func TestSelectQuestions_AllNew(t *testing.T) {
	all := []question.Question{
		makeQuestion("q1", "Elixir", "beginner"),
		makeQuestion("q2", "Elixir", "beginner"),
		makeQuestion("q3", "Elixir", "beginner"),
	}
	store := make(Store)
	today := time.Now()

	selected := SelectQuestions(all, store, 2, nil, "", today)
	if len(selected) != 2 {
		t.Errorf("expected 2 questions, got %d", len(selected))
	}
}

func TestSelectQuestions_DueBucket(t *testing.T) {
	all := []question.Question{
		makeQuestion("q1", "Elixir", "beginner"),
		makeQuestion("q2", "Elixir", "beginner"),
	}
	today := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	store := make(Store)
	// q1 is due today
	store["q1"] = &Record{
		TimesSeen:    1,
		NextDue:      "2026-03-02",
		EaseFactor:   2.5,
		IntervalDays: 1,
	}
	// q2 is due in the future
	store["q2"] = &Record{
		TimesSeen:    1,
		NextDue:      "2026-03-10",
		EaseFactor:   2.5,
		IntervalDays: 8,
	}

	selected := SelectQuestions(all, store, 10, nil, "", today)
	if len(selected) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(selected))
	}
	// q1 should appear (bucket1) before q2 (bucket3)
	if selected[0].ID != "q1" {
		t.Errorf("expected q1 first (due today), got %s", selected[0].ID)
	}
}

func TestSelectQuestions_TopicFilter(t *testing.T) {
	all := []question.Question{
		makeQuestion("q1", "Elixir", "beginner"),
		makeQuestion("q2", "Go", "beginner"),
		makeQuestion("q3", "Elixir", "beginner"),
	}
	store := make(Store)
	today := time.Now()

	selected := SelectQuestions(all, store, 10, []string{"Elixir"}, "", today)
	for _, q := range selected {
		if q.Topic != "Elixir" {
			t.Errorf("expected only Elixir questions, got topic %q", q.Topic)
		}
	}
	if len(selected) != 2 {
		t.Errorf("expected 2 Elixir questions, got %d", len(selected))
	}
}

func TestSelectQuestions_DifficultyFilter(t *testing.T) {
	all := []question.Question{
		makeQuestion("q1", "Elixir", "beginner"),
		makeQuestion("q2", "Elixir", "intermediate"),
		makeQuestion("q3", "Elixir", "beginner"),
	}
	store := make(Store)
	today := time.Now()

	selected := SelectQuestions(all, store, 10, nil, "beginner", today)
	for _, q := range selected {
		if q.Difficulty != "beginner" {
			t.Errorf("expected only beginner questions, got difficulty %q", q.Difficulty)
		}
	}
	if len(selected) != 2 {
		t.Errorf("expected 2 beginner questions, got %d", len(selected))
	}
}

func TestSelectQuestions_CountCap(t *testing.T) {
	var all []question.Question
	for i := 0; i < 20; i++ {
		all = append(all, makeQuestion(
			"q"+string(rune('a'+i)), "Elixir", "beginner",
		))
	}
	store := make(Store)
	today := time.Now()

	selected := SelectQuestions(all, store, 5, nil, "", today)
	if len(selected) != 5 {
		t.Errorf("expected 5 questions (count cap), got %d", len(selected))
	}
}
