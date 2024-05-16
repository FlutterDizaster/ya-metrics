package memory

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Метод запускающий сервис бекапа.
func (ms *MetricStorage) Start(ctx context.Context) error {
	slog.Debug("Start backup service")
	defer slog.Debug("Backup service successfully stopped")
	ticker := &time.Ticker{
		C: make(<-chan time.Time),
	}
	if ms.storeInterval != 0 {
		ticker = time.NewTicker(time.Duration(ms.storeInterval) * time.Second)
	}

	var wg sync.WaitGroup

	for {
		select {
		// Grasefull Shutdown
		case <-ctx.Done():
			ticker.Stop()
			if ms.awaiting.Load() {
				ms.cond.Broadcast()
			} else {
				ms.backup(true)
			}
			wg.Wait()
			return nil
		case <-ticker.C:
			if !ms.awaiting.Load() {
				ms.awaiting.Store(true)
				wg.Add(1)
				go func() {
					ms.backup(false)
					wg.Done()
				}()
			}
		default:
			if ms.storeInterval == 0 {
				if !ms.awaiting.Load() {
					ms.awaiting.Store(true)
					wg.Add(1)
					go func() {
						ms.backup(false)
						wg.Done()
					}()
				}
			}
		}
	}
}

// Метод сохраняющий метрики на диск.
func (ms *MetricStorage) backup(skipWait bool) {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	if !skipWait {
		ms.cond.Wait()
	}

	slog.Debug("Creating backup", slog.String("destination", ms.fileStoragePath))

	err := ms.saveToFile()
	if err != nil {
		slog.Error("backup error", "error", err)
	}

	slog.Debug("Backup created")
	ms.awaiting.Store(false)
}
