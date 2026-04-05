package algorithms

import (
	"fmt"
	"sync/atomic"
	"time"
)

type retryEntry struct {
	id      int
	retries int
}

// FrontierScanner discovers resources using two independent worker pools:
// frontier workers continuously scan forward from startID, while retry workers
// handle failed IDs in the background. maxRequestsPerSec limits retry fetch
// calls per second (0 = unlimited).
func FrontierScanner(startID int, maxWorkers int, maxRetries int, windowSize int, maxRequestsPerSec int, fetch FetchFunc) {
	if maxWorkers < 2 {
		maxWorkers = 2
	}
	if maxRetries <= 0 {
		maxRetries = 1
	}

	frontierWorkers := maxWorkers / 2
	retryWorkers := maxWorkers - frontierWorkers

	retryQueue := make(chan retryEntry, 1<<20)

	var retryLimiter <-chan time.Time
	if maxRequestsPerSec > 0 {
		rt := time.NewTicker(time.Second / time.Duration(maxRequestsPerSec))
		defer rt.Stop()
		retryLimiter = rt.C
	}

	var frontierCursor atomic.Int64
	frontierCursor.Store(int64(startID))

	// Debug ticker — prints state every 5 seconds.
	go func() {
		for range time.NewTicker(5 * time.Second).C {
			fmt.Printf("[FrontierScanner] frontier=%d | retryQueue=%d\n",
				int(frontierCursor.Load()), len(retryQueue))
		}
	}()

	// Frontier workers — scan forward continuously, never throttled.
	for w := 0; w < frontierWorkers; w++ {
		go func() {
			for {
				id := int(frontierCursor.Add(1) - 1)
				if fetch(id) != nil {
					select {
					case retryQueue <- retryEntry{id: id}:
					default:
					}
				}
			}
		}()
	}

	// Retry workers — drain the queue, O(1) per item.
	done := make(chan struct{})
	for w := 0; w < retryWorkers; w++ {
		go func() {
			for entry := range retryQueue {
				if retryLimiter != nil {
					<-retryLimiter
				}
				if fetch(entry.id) == nil {
					continue
				}
				entry.retries++
				if entry.retries < maxRetries {
					select {
					case retryQueue <- entry:
					default:
					}
				}
			}
			close(done)
		}()
	}

	<-done
}
