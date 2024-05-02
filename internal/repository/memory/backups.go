package memory

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func (ms *MetricStorage) StartBackups(ctx context.Context) {
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
		case <-ctx.Done():
			ticker.Stop()
			if ms.isAwaiting() {
				ms.cond.Broadcast()
			} else {
				ms.backup(true)
			}
			wg.Wait()
			return
		case <-ticker.C:
			if !ms.isAwaiting() {
				ms.setAwaiting(true)
				wg.Add(1)
				go func() {
					ms.backup(false)
					wg.Done()
				}()
			}
		default:
			if ms.storeInterval == 0 {
				if !ms.isAwaiting() {
					ms.setAwaiting(true)
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

func (ms *MetricStorage) isAwaiting() bool {
	ms.awmtx.Lock()
	defer ms.awmtx.Unlock()
	return ms.awaiting
}

func (ms *MetricStorage) setAwaiting(is bool) {
	ms.awmtx.Lock()
	ms.awaiting = is
	ms.awmtx.Unlock()
}

func (ms *MetricStorage) backup(skipWait bool) {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	// slog.Debug("Waiting metrics for backup")
	if !skipWait {
		// slog.Debug("Waiting new data")
		ms.cond.Wait()
	}

	slog.Debug("Creating backup", slog.String("destination", ms.fileStoragePath))

	err := ms.saveToFile()
	if err != nil {
		slog.Error("backup error", "error", err)
	}

	slog.Debug("Backup created")
	ms.setAwaiting(false)
}
