package utils

import "sync"

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

type WaitGroupUtil struct {
	WaitGroup *sync.WaitGroup
}

func (k *WaitGroupUtil) Wrapper(f func()) {
	k.WaitGroup.Add(1)
	go func() {
		defer k.WaitGroup.Done()
		f()
	}()
}
