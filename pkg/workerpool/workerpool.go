package workerpool

import (
	"errors"
	"sync"
)

type WorkerFunc func()

type WorkerPool struct {
	maxWorkers int
	workQ      chan WorkerFunc
	wg         sync.WaitGroup
}

func New(maxWorkers int) *WorkerPool {
	wp := WorkerPool{
		maxWorkers: maxWorkers,
		workQ:      make(chan WorkerFunc, maxWorkers),
		wg:         sync.WaitGroup{},
	}

	go wp.controller()

	return &wp
}

func (wp *WorkerPool) Do(fn WorkerFunc) error {
	wp.workQ <- fn

	if err := recover(); err != nil {
		return errors.New("worker pool closed")
	}
	return nil
}

func (wp *WorkerPool) Close() {
	close(wp.workQ)
	wp.wg.Wait()
}

func (wp *WorkerPool) controller() {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			wp.worker()
			wp.wg.Done()
		}()
	}
}

func (wp *WorkerPool) worker() {
	for payload := range wp.workQ {
		payload()
	}
}
