package ratelimit

import "time"

type RateLimit interface {
	Allow() bool
	Take() time.Time
}
