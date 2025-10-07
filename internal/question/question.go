package question

import (
	"fmt"
	"io/fs"
	"sort"

	"gopkg.in/yaml.v3"
)

type QuestionType string

const (
	SingleChoice   QuestionType = "single_choice"
	MultipleChoice QuestionType = "multiple_choice"
	CodeSnippet    QuestionType = "code_snippet"
)

type Option struct {
	ID      string `yaml:"id"`
	Text    string `yaml:"text"`
	Correct bool   `yaml:"correct"`
}

type Question struct {
	ID          string       `yaml:"id"`
	Version     string       `yaml:"version"`
	Topic       string       `yaml:"topic"`
	Subtopic    string       `yaml:"subtopic"`
	Difficulty  string       `yaml:"difficulty"`
	DocURL      string       `yaml:"doc_url"`
	Type        QuestionType `yaml:"type"`
	Question    string       `yaml:"question"`
	Code        string       `yaml:"code"`
	Options     []Option     `yaml:"options"`
	Explanation string       `yaml:"explanation"`
	Tags        []string     `yaml:"tags"`
}

func (q *Question) Validate() error {
	if q.ID == "" {
		return fmt.Errorf("question missing ID")
	}
	if q.Topic == "" {
		return fmt.Errorf("question %s: missing topic", q.ID)
	}
	if q.Question == "" {
		return fmt.Errorf("question %s: missing question text", q.ID)
	}
	if len(q.Options) < 2 {
		return fmt.Errorf("question %s: needs at least 2 options", q.ID)
	}
	switch q.Type {
	case SingleChoice, CodeSnippet:
		correct := 0
		for _, o := range q.Options {
			if o.Correct {
				correct++
			}
		}
		if correct != 1 {
			return fmt.Errorf("question %s: single_choice/code_snippet must have exactly 1 correct option, found %d", q.ID, correct)
		}
	case MultipleChoice:
		correct := 0
		for _, o := range q.Options {
			if o.Correct {
				correct++
			}
		}
		if correct < 1 {
			return fmt.Errorf("question %s: multiple_choice must have at least 1 correct option", q.ID)
		}
	default:
		return fmt.Errorf("question %s: unknown type %q", q.ID, q.Type)
	}
	return nil
}

// LoadAll walks the embedded FS, parses all YAML files, validates questions,
// and returns the questions plus a sorted list of unique topic names.
func LoadAll(fsys fs.FS) ([]Question, []string, error) {
	var all []Question
	seen := map[string]bool{}
	topicSet := map[string]bool{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Only process .yaml/.yml files
		if len(path) < 5 {
			return nil
		}
		ext := path[len(path)-5:]
		if ext != ".yaml" && path[len(path)-4:] != ".yml" {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		var questions []Question
		if err := yaml.Unmarshal(data, &questions); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		for i := range questions {
			q := &questions[i]
			if err := q.Validate(); err != nil {
				return fmt.Errorf("file %s: %w", path, err)
			}
			if seen[q.ID] {
				return fmt.Errorf("file %s: duplicate question ID %q", path, q.ID)
			}
			seen[q.ID] = true
			topicSet[q.Topic] = true
			all = append(all, *q)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	topics := make([]string, 0, len(topicSet))
	for t := range topicSet {
		topics = append(topics, t)
	}
	sort.Strings(topics)

	return all, topics, nil
}
