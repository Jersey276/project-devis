package controllers

import (
	"sync"
	"time"
)

type slidingWindowLimiter struct {
	mu      sync.Mutex
	entries map[string][]time.Time
}

func newSlidingWindowLimiter() *slidingWindowLimiter {
	return &slidingWindowLimiter{entries: make(map[string][]time.Time)}
}

func (l *slidingWindowLimiter) Allow(key string, limit int, window time.Duration, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	min := now.Add(-window)
	hits := l.entries[key]
	filtered := hits[:0]
	for _, ts := range hits {
		if ts.After(min) {
			filtered = append(filtered, ts)
		}
	}

	if len(filtered) >= limit {
		l.entries[key] = filtered
		return false
	}

	filtered = append(filtered, now)
	l.entries[key] = filtered
	return true
}
