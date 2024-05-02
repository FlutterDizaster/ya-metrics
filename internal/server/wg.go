package server

import (
	"log/slog"
	"sync"
)

// TODO: for testing
type CustomWG struct {
	sync.WaitGroup
	count int
	mtx   sync.Mutex
}

func (cwg *CustomWG) Add(count int, name ...string) {
	n := ""
	if len(name) > 0 {
		n = name[0]
	}
	cwg.mtx.Lock()
	cwg.count += count
	slog.Debug("WaitGroup added", "count", cwg.count, "name", n)
	cwg.WaitGroup.Add(count)
	cwg.mtx.Unlock()
}

func (cwg *CustomWG) Done(name ...string) {
	n := ""
	if len(name) > 0 {
		n = name[0]
	}
	cwg.mtx.Lock()
	cwg.count--
	slog.Debug("WaitGroup done", "count", cwg.count, "name", n)
	cwg.WaitGroup.Done()
	cwg.mtx.Unlock()
}
