package question

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

const validYAML = `
- id: test-001
  version: "1"
  topic: Test
  subtopic: Basics
  difficulty: beginner
  type: single_choice
  question: "Which is correct?"
  options:
    - id: a
      text: "Right"
      correct: true
    - id: b
      text: "Wrong"
      correct: false
  explanation: "A is right."
`

const multipleYAML = `
- id: multi-001
  version: "1"
  topic: Test
  subtopic: Multi
  difficulty: beginner
  type: multiple_choice
  question: "Which are correct?"
  options:
    - id: a
      text: "Right 1"
      correct: true
    - id: b
      text: "Right 2"
      correct: true
    - id: c
      text: "Wrong"
      correct: false
`

const codeYAML = `
- id: code-001
  version: "1"
  topic: Test
  subtopic: Code
  difficulty: intermediate
  type: code_snippet
  question: "What does this output?"
  code: "IO.puts(1)"
  options:
    - id: a
      text: "1"
      correct: true
    - id: b
      text: "2"
      correct: false
`

func makeFS(files map[string]string) fs.FS {
	m := fstest.MapFS{}
	for name, content := range files {
		m[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return m
}

func TestLoadAll_SingleFile(t *testing.T) {
	fsys := makeFS(map[string]string{"questions/elixir/basics.yaml": validYAML})
	qs, topics, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 1 {
		t.Fatalf("expected 1 question, got %d", len(qs))
	}
	if qs[0].ID != "test-001" {
		t.Errorf("expected ID test-001, got %q", qs[0].ID)
	}
	if len(topics) != 1 || topics[0] != "Test" {
		t.Errorf("expected topics [Test], got %v", topics)
	}
}

func TestLoadAll_DuplicateID(t *testing.T) {
	fsys := makeFS(map[string]string{
		"a.yaml": validYAML,
		"b.yaml": validYAML,
	})
	_, _, err := LoadAll(fsys)
	if err == nil {
		t.Fatal("expected duplicate ID error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention duplicate, got: %v", err)
	}
}

func TestLoadAll_TopicsSorted(t *testing.T) {
	yaml2 := strings.ReplaceAll(validYAML, "test-001", "test-002")
	yaml2 = strings.ReplaceAll(yaml2, "topic: Test", "topic: Alpha")
	fsys := makeFS(map[string]string{
		"a.yaml": validYAML,
		"b.yaml": yaml2,
	})
	_, topics, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
	if topics[0] != "Alpha" || topics[1] != "Test" {
		t.Errorf("expected [Alpha Test], got %v", topics)
	}
}

func TestValidate_SingleChoice_NoCorrect(t *testing.T) {
	q := Question{
		ID:       "x",
		Topic:    "T",
		Question: "Q?",
		Type:     SingleChoice,
		Options: []Option{
			{ID: "a", Text: "A", Correct: false},
			{ID: "b", Text: "B", Correct: false},
		},
	}
	if err := q.Validate(); err == nil {
		t.Error("expected error for no correct option")
	}
}

func TestValidate_MultipleChoice_OK(t *testing.T) {
	fsys := makeFS(map[string]string{"q.yaml": multipleYAML})
	qs, _, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(qs) != 1 {
		t.Fatalf("expected 1, got %d", len(qs))
	}
}

func TestValidate_CodeSnippet(t *testing.T) {
	fsys := makeFS(map[string]string{"q.yaml": codeYAML})
	qs, _, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if qs[0].Type != CodeSnippet {
		t.Errorf("expected code_snippet type")
	}
	if qs[0].Code == "" {
		t.Errorf("expected non-empty code")
	}
}

func TestValidate_MissingID(t *testing.T) {
	q := Question{Topic: "T", Question: "Q?", Type: SingleChoice}
	if err := q.Validate(); err == nil {
		t.Error("expected error for missing ID")
	}
}
