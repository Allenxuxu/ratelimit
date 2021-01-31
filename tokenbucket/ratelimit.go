package tokenbucket

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Allenxuxu/ratelimit"
	uAtomic "go.uber.org/atomic"
)

type state struct {
	last            time.Time
	availableTokens int64
}

type limiter struct {
	cap      int64
	perToken *uAtomic.Int64
	state    unsafe.Pointer

	opts ratelimit.Options

	time clock // for unit test
}

func New(rate, cap int64, opts ...ratelimit.Option) ratelimit.RateLimit {
	return newLimit(rate, cap, opts...)
}

func newLimit(rate, cap int64, opts ...ratelimit.Option) *limiter {
	options := ratelimit.Options{
		Per: time.Second,
	}

	for _, o := range opts {
		o(&options)
	}

	l := &limiter{
		cap:      cap,
		opts:     options,
		time:     &realClock{},
		perToken: uAtomic.NewInt64(0),
	}

	l.perToken.Store(int64(l.opts.Per / time.Duration(rate)))

	s := state{
		last:            l.time.Now(),
		availableTokens: cap,
	}
	atomic.StorePointer(&l.state, unsafe.Pointer(&s))

	if l.opts.DynamicLimitLoop != nil {
		go l.opts.DynamicLimitLoop(l.perToken, rate)
	}

	return l
}

func (l *limiter) Take() time.Time {
	newState := state{}
	taken := false
	for !taken {
		previousStatePointer := atomic.LoadPointer(&l.state)
		oldState := (*state)(previousStatePointer)

		last := oldState.last
		now := l.time.Now()

		newState.last = now
		newState.availableTokens = min(l.cap, oldState.availableTokens+int64(now.Sub(last)/time.Duration(l.perToken.Load())))

		if newState.availableTokens > 0 {
			newState.availableTokens--
			taken = atomic.CompareAndSwapPointer(&l.state, previousStatePointer, unsafe.Pointer(&newState))
		}
		if !taken {
			time.Sleep(time.Duration(l.perToken.Load()))
		}
	}

	return newState.last
}

func (l *limiter) Allow() bool {
	previousStatePointer := atomic.LoadPointer(&l.state)
	oldState := (*state)(previousStatePointer)

	last := oldState.last
	now := l.time.Now()

	newState := state{
		last:            now,
		availableTokens: min(l.cap, oldState.availableTokens+int64(now.Sub(last)/time.Duration(l.perToken.Load()))),
	}

	if newState.availableTokens > 0 {
		newState.availableTokens--
		return atomic.CompareAndSwapPointer(&l.state, previousStatePointer, unsafe.Pointer(&newState))
	}

	return false
}

func (l *limiter) availableTokens() int64 {
	previousStatePointer := atomic.LoadPointer(&l.state)
	oldState := (*state)(previousStatePointer)

	return oldState.availableTokens
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}
