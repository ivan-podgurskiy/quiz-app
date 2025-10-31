package session

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"sync"

	"quiz-app/internal/question"
)

// Answer records the user's response to a single question.
type Answer struct {
	QuestionID  string
	SelectedIDs []string
	Correct     bool
}

// Session represents a single quiz session.
type Session struct {
	ID           string
	Questions    []question.Question
	CurrentIndex int
	Answers      []Answer
	ShuffledOpts map[string][]question.Option // pre-shuffled options per question ID
}

// Store is a thread-safe in-memory store for sessions.
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewStore creates a new session Store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

// Create creates a new session with the given questions and pre-shuffled options.
func (s *Store) Create(questions []question.Question) *Session {
	id := newUUID()

	shuffled := make(map[string][]question.Option, len(questions))
	for _, q := range questions {
		opts := make([]question.Option, len(q.Options))
		copy(opts, q.Options)
		rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
		shuffled[q.ID] = opts
	}

	sess := &Session{
		ID:           id,
		Questions:    questions,
		CurrentIndex: 0,
		Answers:      []Answer{},
		ShuffledOpts: shuffled,
	}

	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()

	return sess
}

// Get retrieves a session by ID, returning nil if not found.
func (s *Store) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

// Delete removes a session from the store.
func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// newUUID generates a random 16-byte hex string using crypto/rand.
func newUUID() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}
