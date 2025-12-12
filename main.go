package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"

	"quiz-app/internal/progress"
	"quiz-app/internal/question"
	"quiz-app/internal/server"
	"quiz-app/internal/session"
)

//go:embed templates
var templateFS embed.FS

//go:embed questions
var questionsFS embed.FS

func main() {
	questions, topics, err := question.LoadAll(questionsFS)
	if err != nil {
		log.Fatalf("loading questions: %v", err)
	}
	log.Printf("Loaded %d questions across %d topic(s)", len(questions), len(topics))

	prog, err := progress.Load()
	if err != nil {
		log.Printf("Warning: could not load progress: %v — starting fresh", err)
		prog = make(progress.Store)
	}

	funcMap := template.FuncMap{
		"percent": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a * 100 / b
		},
		"optionClass": func(correct, selected bool) string {
			if correct && selected {
				return "correct"
			}
			if correct {
				return "highlight"
			}
			if selected {
				return "wrong"
			}
			return ""
		},
		"scoreClass": func(pct int) string {
			if pct >= 60 {
				return "pass"
			}
			return "fail"
		},
		"topicBarColor": func(pct int) string {
			switch {
			case pct >= 80:
				return "#56d364"
			case pct >= 60:
				return "#e3b341"
			default:
				return "#ff7b72"
			}
		},
		"safe":    func(s string) template.HTML { return template.HTML(s) },
		"add":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"sprintf": fmt.Sprintf,
	}

	// Build one template set per page: layout + that page only.
	// This prevents {{define "content"}} blocks from different pages
	// from overwriting each other when parsed into the same set.
	pages := []string{"home.html", "question.html", "feedback.html", "summary.html"}
	templates := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		t := template.Must(
			template.New("").Funcs(funcMap).ParseFS(
				templateFS,
				"templates/layout.html",
				"templates/"+page,
			),
		)
		templates[page] = t
	}

	dep := &server.Dependencies{
		Questions:     questions,
		Topics:        topics,
		ProgressStore: prog,
		SessionStore:  session.NewStore(),
		Templates:     templates,
	}

	r := server.New(dep)

	log.Println("Starting Quiz App on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
