package workerpool

import (
	"errors"
	"sync"
)

// WorkerFunc - функция, выполняемая в отдельном потоке.
type WorkerFunc func()

// WorkerPool - структура для запуска одновременно работащих воркеров.
// Максимальное кол-во воркеров задается при создании через New().
type WorkerPool struct {
	maxWorkers int
	workQ      chan WorkerFunc
	wg         sync.WaitGroup
}

// New - создание нового объекта WorkerPool.
// maxWorkers - максимальное кол-во одновременно работащих воркеров.
func New(maxWorkers int) *WorkerPool {
	wp := WorkerPool{
		maxWorkers: maxWorkers,
		workQ:      make(chan WorkerFunc, maxWorkers),
		wg:         sync.WaitGroup{},
	}

	go wp.controller()

	return &wp
}

// Do - запуск функции fn в отдельном потоке.
// Блокирует текущий поток, если уже запущено максимальное кол-во воркеров.
// Поток разблокируется, когда новый воркер добавится в очередь.
// Возвращает ошибку, если WorkerPool закрыт.
func (wp *WorkerPool) Do(fn WorkerFunc) error {
	wp.workQ <- fn

	if err := recover(); err != nil {
		return errors.New("worker pool closed")
	}
	return nil
}

// Close - закрытие WorkerPool.
// После вызова добавление новых воркеров через Do() будет невозможно.
// Блокирует текущий поток до завершения всех воркеров.
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
