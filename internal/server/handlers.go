package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"quiz-app/internal/progress"
	"quiz-app/internal/question"
	"quiz-app/internal/session"
)

type handlers struct {
	dep *Dependencies
}

// render executes the named page template (which includes layout) directly into
// the response writer. Each page has its own template.Template set so that
// {{define "content"}} blocks don't overwrite each other.
func (h *handlers) render(c *gin.Context, status int, page string, data any) {
	tmpl, ok := h.dep.Templates[page]
	if !ok {
		log.Printf("render: unknown template %q", page)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	if err := tmpl.ExecuteTemplate(c.Writer, "layout", data); err != nil {
		log.Printf("template error (%s): %v", page, err)
	}
}

// homeData is passed to templates/home.html
type homeData struct {
	Topics []string
	Error  string
}

// HandlerHome renders the topic/difficulty/count selection form.
func (h *handlers) HandlerHome(c *gin.Context) {
	h.render(c, http.StatusOK, "home.html", homeData{Topics: h.dep.Topics})
}

// HandlerStart reads form values, selects questions, creates a session, and redirects.
func (h *handlers) HandlerStart(c *gin.Context) {
	// Parse topics (multi-value form field)
	topics := c.PostFormArray("topics")
	difficulty := c.PostForm("difficulty")
	countStr := c.PostForm("count")

	count := 10
	if n, err := strconv.Atoi(countStr); err == nil && n > 0 {
		count = n
	}

	selected := progress.SelectQuestions(
		h.dep.Questions,
		h.dep.ProgressStore,
		count,
		topics,
		difficulty,
		time.Now(),
	)

	if len(selected) == 0 {
		h.render(c, http.StatusOK, "home.html", homeData{
			Topics: h.dep.Topics,
			Error:  "No questions match the selected filters. Try different options.",
		})
		return
	}

	sess := h.dep.SessionStore.Create(selected)
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/quiz/%s/question", sess.ID))
}

// questionData is passed to templates/question.html
type questionData struct {
	SessionID  string
	Question   question.Question
	Options    []question.Option
	Index      int // 1-based
	Total      int
	IsCode     bool
	IsMultiple bool
	LangClass  string // highlight.js language class, e.g. "language-go"
}

// HandlerQuestion shows the current question, or redirects to summary if done.
func (h *handlers) HandlerQuestion(c *gin.Context) {
	sid := c.Param("sid")
	sess := h.dep.SessionStore.Get(sid)
	if sess == nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if sess.CurrentIndex >= len(sess.Questions) {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/quiz/%s/summary", sid))
		return
	}

	q := sess.Questions[sess.CurrentIndex]
	opts := sess.ShuffledOpts[q.ID]

	h.render(c, http.StatusOK, "question.html", questionData{
		SessionID:  sid,
		Question:   q,
		Options:    opts,
		Index:      sess.CurrentIndex + 1,
		Total:      len(sess.Questions),
		IsCode:     q.Type == question.CodeSnippet,
		IsMultiple: q.Type == question.MultipleChoice,
		LangClass:  langClass(q.Topic),
	})
}

// feedbackData is passed to templates/feedback.html
type feedbackData struct {
	SessionID   string
	Question    question.Question
	Options     []question.Option
	SelectedIDs map[string]bool
	Correct     bool
	IsLast      bool
	IsCode      bool
	LangClass   string
}

// HandlerAnswer evaluates the submitted answer, updates SM-2 progress, and renders feedback.
func (h *handlers) HandlerAnswer(c *gin.Context) {
	sid := c.Param("sid")
	sess := h.dep.SessionStore.Get(sid)
	if sess == nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if sess.CurrentIndex >= len(sess.Questions) {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/quiz/%s/summary", sid))
		return
	}

	q := sess.Questions[sess.CurrentIndex]

	// Collect submitted option IDs
	submitted := c.PostFormArray("options")
	submittedSet := make(map[string]bool, len(submitted))
	for _, id := range submitted {
		submittedSet[id] = true
	}

	// Determine correctness
	correct := isCorrect(q, submittedSet)

	// Record answer
	sess.Answers = append(sess.Answers, session.Answer{
		QuestionID:  q.ID,
		SelectedIDs: submitted,
		Correct:     correct,
	})
	sess.CurrentIndex++

	// Update SM-2 progress
	rec := h.dep.ProgressStore.Get(q.ID)
	if rec == nil {
		rec = &progress.Record{
			EaseFactor:   2.5,
			IntervalDays: 1.0,
		}
		h.dep.ProgressStore[q.ID] = rec
	}
	rec.Update(correct, time.Now())

	// Save progress (best-effort)
	_ = h.dep.ProgressStore.Save()

	opts := sess.ShuffledOpts[q.ID]
	isLast := sess.CurrentIndex >= len(sess.Questions)

	h.render(c, http.StatusOK, "feedback.html", feedbackData{
		SessionID:   sid,
		Question:    q,
		Options:     opts,
		SelectedIDs: submittedSet,
		Correct:     correct,
		IsLast:      isLast,
		IsCode:      q.Type == question.CodeSnippet,
		LangClass:   langClass(q.Topic),
	})
}

// langClass returns the highlight.js CSS class for a given topic name.
func langClass(topic string) string {
	switch topic {
	case "Go":
		return "language-go"
	case "JavaScript":
		return "language-javascript"
	case "Python":
		return "language-python"
	case "AWS":
		return "language-json"
	default:
		return "language-elixir"
	}
}

// isCorrect checks whether the submitted set of option IDs is exactly correct.
func isCorrect(q question.Question, submitted map[string]bool) bool {
	correctSet := make(map[string]bool)
	for _, o := range q.Options {
		if o.Correct {
			correctSet[o.ID] = true
		}
	}
	if len(submitted) != len(correctSet) {
		return false
	}
	for id := range correctSet {
		if !submitted[id] {
			return false
		}
	}
	return true
}

// topicStat holds per-topic stats for the summary page.
type topicStat struct {
	Topic   string
	Correct int
	Total   int
	Percent int
}

// summaryData is passed to templates/summary.html
type summaryData struct {
	Total      int
	Correct    int
	Percent    int
	TopicStats []topicStat
	WeakTopics []string
}

// HandlerSummary computes session stats, renders the summary, and deletes the session.
func (h *handlers) HandlerSummary(c *gin.Context) {
	sid := c.Param("sid")
	sess := h.dep.SessionStore.Get(sid)
	if sess == nil {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	// Build question lookup
	qMap := make(map[string]question.Question, len(sess.Questions))
	for _, q := range sess.Questions {
		qMap[q.ID] = q
	}

	// Aggregate stats
	topicCorrect := map[string]int{}
	topicTotal := map[string]int{}
	totalCorrect := 0

	for _, ans := range sess.Answers {
		q, ok := qMap[ans.QuestionID]
		if !ok {
			continue
		}
		topicTotal[q.Topic]++
		if ans.Correct {
			totalCorrect++
			topicCorrect[q.Topic]++
		}
	}

	total := len(sess.Answers)
	percent := 0
	if total > 0 {
		percent = totalCorrect * 100 / total
	}

	var stats []topicStat
	var weakTopics []string
	for topic, tot := range topicTotal {
		cor := topicCorrect[topic]
		pct := 0
		if tot > 0 {
			pct = cor * 100 / tot
		}
		stats = append(stats, topicStat{
			Topic:   topic,
			Correct: cor,
			Total:   tot,
			Percent: pct,
		})
		if pct < 60 {
			weakTopics = append(weakTopics, topic)
		}
	}

	h.dep.SessionStore.Delete(sid)

	h.render(c, http.StatusOK, "summary.html", summaryData{
		Total:      total,
		Correct:    totalCorrect,
		Percent:    percent,
		TopicStats: stats,
		WeakTopics: weakTopics,
	})
}
