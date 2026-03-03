# Quiz App

A self-hosted developer quiz and spaced-repetition training tool. Answer questions, get instant feedback, and let the SM-2 algorithm surface the ones you keep getting wrong.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go) ![License](https://img.shields.io/badge/license-MIT-green?style=flat)

---

## Features

- **Spaced repetition** — SM-2 algorithm prioritises due and unseen questions; questions you miss come back sooner
- **Multiple question types** — single choice, multiple choice, and code snippet (with syntax highlighting)
- **Four topics out of the box** — Elixir (beginner → advanced), Go, JavaScript, Python
- **Per-topic summary** — score breakdown and weak-topic detection after each session
- **Single binary** — templates and question files are embedded; nothing to deploy alongside the binary
- **Progress persistence** — answers saved to `~/.quiz/progress.json` between sessions

## Quick start

```bash
git clone https://github.com/ivan-podgurskiy/quiz-app.git
cd quiz-app
go build -o quiz-app .
./quiz-app
```

Then open [http://localhost:8080](http://localhost:8080).

**Requirements:** Go 1.22+

## How it works

### Spaced repetition (SM-2)

Each question has a record in `~/.quiz/progress.json` tracking ease factor, interval, and due date. Questions are drawn from three priority buckets:

| Bucket | Condition | Priority |
|--------|-----------|----------|
| 1 | `next_due ≤ today` | High — shown first |
| 2 | `times_seen == 0` | Medium — new questions |
| 3 | everything else | Low — filler |

A correct answer increases the interval (`interval × ease_factor`) and nudges ease up (max 2.5). A wrong answer resets the interval to 1 day and drops ease (floor 1.3).

### Question format

Questions live in `questions/<topic>/<file>.yaml`. Add a new YAML file to add a topic — no code changes needed.

```yaml
- id: go-basics-001
  version: "1"
  topic: Go
  subtopic: Goroutines
  difficulty: beginner          # beginner | intermediate | advanced
  type: single_choice           # single_choice | multiple_choice | code_snippet
  question: "What keyword starts a goroutine?"
  code: |                       # only for code_snippet type
    go func() { ... }()
  options:
    - id: a
      text: "go func()"
      correct: true
    - id: b
      text: "async func()"
      correct: false
  explanation: "The go keyword launches a goroutine."
  doc_url: "https://go.dev/tour/concurrency/1"
  tags: [goroutines, concurrency]
```

**Rules enforced at startup:**
- `single_choice` and `code_snippet` must have exactly one correct option
- `multiple_choice` must have at least one correct option
- Question IDs must be unique across all files

## Project structure

```
quiz-app/
├── main.go                      # embed FS, template funcmap, wire deps, start server
├── internal/
│   ├── question/                # Question/Option structs, LoadAll, Validate
│   ├── progress/                # SM-2 Record/Store, Load/Save, SelectQuestions
│   ├── session/                 # in-memory session store (sync.RWMutex, crypto/rand)
│   └── server/                  # Gin routes, all five HTTP handlers
├── templates/                   # html/template files (embedded)
│   ├── layout.html              # base layout, highlight.js CDN, CSS
│   ├── home.html                # topic/difficulty/count form
│   ├── question.html            # question form (radio / checkbox)
│   ├── feedback.html            # correct/wrong, explanation, doc link
│   └── summary.html             # score, per-topic table, weak topics
└── questions/
    ├── elixir/
    │   ├── basics.yaml          # 10 questions — beginner/intermediate
    │   └── advanced.yaml        # 8 questions — OTP, GenServer, ETS, Streams
    ├── go/
    │   └── basics.yaml          # 10 questions — goroutines, channels, closures
    ├── javascript/
    │   └── basics.yaml          # 10 questions — event loop, promises, closures
    └── python/
        └── basics.yaml          # 10 questions — generators, decorators, GIL
```

## Running tests

```bash
go test ./...
```

## Adding questions

1. Create or edit a YAML file under `questions/<topic>/`
2. Follow the format above — pick a unique `id`, set `type`, add at least two options
3. Rebuild: `go build -o quiz-app . && ./quiz-app`

The new topic will appear automatically in the home form.
