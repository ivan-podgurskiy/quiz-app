package server

import (
	"html/template"

	"github.com/gin-gonic/gin"

	"quiz-app/internal/progress"
	"quiz-app/internal/question"
	"quiz-app/internal/session"
)

// Dependencies holds all wired components for the server.
type Dependencies struct {
	Questions     []question.Question
	Topics        []string
	ProgressStore progress.Store
	SessionStore  *session.Store
	Templates     map[string]*template.Template // one set per page
}

// New creates a configured Gin engine with all routes registered.
func New(dep *Dependencies) *gin.Engine {
	r := gin.Default()

	h := &handlers{dep: dep}

	r.GET("/", h.HandlerHome)
	r.POST("/quiz/start", h.HandlerStart)
	r.GET("/quiz/:sid/question", h.HandlerQuestion)
	r.POST("/quiz/:sid/answer", h.HandlerAnswer)
	r.GET("/quiz/:sid/summary", h.HandlerSummary)

	return r
}
