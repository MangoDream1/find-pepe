package scraper

import (
	"sync"

	"github.com/google/uuid"
)

type WaitGroupHelper struct {
	WaitGroup *sync.WaitGroup
}

func (k *WaitGroupHelper) Wrapper(f func()) {
	k.WaitGroup.Add(1)
	go func() {
		defer writeToPanicFile()
		defer k.WaitGroup.Done()
		f()
	}()
}

func createUniqueId() string {
	uuid := uuid.New()
	return uuid.String()
}
