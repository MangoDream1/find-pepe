package limiter

import "sync"

type Limiter struct {
	Total   uint8
	amount  uint8
	waiting uint8
	done    chan bool
	m       sync.Mutex
}

func NewLimiter(total uint8) *Limiter {
	return &Limiter{
		Total: total,
		m:     sync.Mutex{},
		done:  make(chan bool),
	}
}

func (l *Limiter) Add() {
	l.m.Lock()

	if l.amount >= l.Total {
		l.waiting += 1
		<-l.done
	}

	l.amount += 1
	l.m.Unlock()
}

func (l *Limiter) Done() {
	l.amount -= 1

	if l.waiting > 0 {
		l.waiting -= 1
		l.done <- true
	}

}
