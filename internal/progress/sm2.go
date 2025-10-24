package progress

import (
	"math/rand"
	"time"

	"quiz-app/internal/question"
)

const (
	initialEaseFactor   = 2.5
	initialIntervalDays = 1.0
	minEaseFactor       = 1.3
	maxEaseFactor       = 2.5
)

// Update applies the SM-2 algorithm to the record.
// correct=true: interval *= easeFactor (min 1), easeFactor += 0.1 (capped at 2.5)
// correct=false: interval = 1, easeFactor -= 0.2 (floored at 1.3)
func (r *Record) Update(correct bool, today time.Time) {
	r.TimesSeen++
	r.LastSeen = today.Format("2006-01-02")

	if r.EaseFactor == 0 {
		r.EaseFactor = initialEaseFactor
	}
	if r.IntervalDays == 0 {
		r.IntervalDays = initialIntervalDays
	}

	if correct {
		r.TimesCorrect++
		r.IntervalDays = max(1, r.IntervalDays*r.EaseFactor)
		r.EaseFactor = min(maxEaseFactor, r.EaseFactor+0.1)
	} else {
		r.IntervalDays = 1
		r.EaseFactor = max(minEaseFactor, r.EaseFactor-0.2)
	}

	nextDue := today.AddDate(0, 0, int(r.IntervalDays))
	r.NextDue = nextDue.Format("2006-01-02")
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// SelectQuestions picks up to `count` questions using 3-bucket priority:
//
//	Bucket 1 (high):   next_due <= today (due for review)
//	Bucket 2 (medium): times_seen == 0   (new questions)
//	Bucket 3 (low):    all others
//
// Buckets are shuffled independently, then concatenated up to `count`.
func SelectQuestions(
	all []question.Question,
	store Store,
	count int,
	topics []string,
	difficulty string,
	today time.Time,
) []question.Question {
	todayStr := today.Format("2006-01-02")

	topicSet := make(map[string]bool, len(topics))
	for _, t := range topics {
		topicSet[t] = true
	}

	var bucket1, bucket2, bucket3 []question.Question

	for _, q := range all {
		if len(topicSet) > 0 && !topicSet[q.Topic] {
			continue
		}
		if difficulty != "" && q.Difficulty != difficulty {
			continue
		}

		rec := store.Get(q.ID)
		if rec == nil || rec.TimesSeen == 0 {
			bucket2 = append(bucket2, q)
		} else if rec.NextDue <= todayStr {
			bucket1 = append(bucket1, q)
		} else {
			bucket3 = append(bucket3, q)
		}
	}

	shuffle := func(qs []question.Question) {
		rand.Shuffle(len(qs), func(i, j int) { qs[i], qs[j] = qs[j], qs[i] })
	}

	shuffle(bucket1)
	shuffle(bucket2)
	shuffle(bucket3)

	combined := append(bucket1, append(bucket2, bucket3...)...)
	if len(combined) > count {
		combined = combined[:count]
	}
	return combined
}
