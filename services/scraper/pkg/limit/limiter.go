package limit

import "sync"

type Limiter interface {
	Add()
	Done()
}

type limiter struct {
	Total   uint8
	amount  uint8
	waiting uint8
	done    chan bool
	m       sync.Mutex
}

func (l *limiter) Add() {
	l.m.Lock()

	if l.amount >= l.Total {
		l.waiting += 1
		<-l.done
	}

	l.amount += 1
	l.m.Unlock()
}

func (l *limiter) Done() {
	l.amount -= 1

	if l.waiting > 0 {
		l.waiting -= 1
		l.done <- true
	}
}

type noopLimiter struct{}

func (l *noopLimiter) Add()  {}
func (l *noopLimiter) Done() {}

func NewLimiter(total uint8) Limiter {
	// 0 is uncapped
	if total == 0 {
		return &noopLimiter{}
	}

	return &limiter{
		Total: total,
		m:     sync.Mutex{},
		done:  make(chan bool),
	}
}
