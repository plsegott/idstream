package algorithms

import (
	"sync"
	"sync/atomic"
	"time"
)

const frontierScannerTickInterval = 1 * time.Second

// FrontierScanner discovers resources using two independent worker pools:
// frontier workers continuously scan ahead of the highest known live ID,
// while retry workers handle failed IDs in the background.
// maxRequestsPerSec limits total fetch calls per second (0 = unlimited).
func FrontierScanner(startID int, maxWorkers int, maxRetries int, windowSize int, maxRequestsPerSec int, fetch FetchFunc) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	if maxRetries <= 0 {
		maxRetries = 1
	}
	if windowSize <= 0 {
		windowSize = 1000
	}

	frontierWorkers := maxWorkers / 2
	if frontierWorkers < 1 {
		frontierWorkers = 1
	}
	retryWorkers := maxWorkers - frontierWorkers

	type pendingEntry struct {
		retries int
	}

	var mu sync.Mutex
	highWater := startID - 1
	nextFrontier := startID
	pending := map[int]pendingEntry{}

	var limiter <-chan time.Time
	if maxRequestsPerSec > 0 {
		ticker := time.NewTicker(time.Second / time.Duration(maxRequestsPerSec))
		defer ticker.Stop()
		limiter = ticker.C
	}

	tryFetch := func(id int) bool {
		if limiter != nil {
			<-limiter
		}
		return fetch(id) == nil
	}

	// Frontier workers — always advance highWater forward.
	var frontierWg sync.WaitGroup
	var frontierCursor atomic.Int64
	frontierCursor.Store(int64(startID))

	for w := 0; w < frontierWorkers; w++ {
		frontierWg.Add(1)
		go func() {
			defer frontierWg.Done()
			for {
				id := int(frontierCursor.Add(1) - 1)

				mu.Lock()
				if id > highWater+windowSize {
					// Caught up to window edge, wait for highWater to advance.
					mu.Unlock()
					time.Sleep(frontierScannerTickInterval)
					continue
				}
				mu.Unlock()

				ok := tryFetch(id)

				mu.Lock()
				if ok {
					if id > highWater {
						highWater = id
					}
					delete(pending, id)
				} else {
					e := pending[id]
					e.retries++
					if e.retries < maxRetries {
						pending[id] = e
					} else {
						delete(pending, id)
					}
				}
				// Keep nextFrontier ahead of highWater.
				if nextFrontier <= highWater {
					nextFrontier = highWater + 1
					frontierCursor.Store(int64(nextFrontier))
				}
				mu.Unlock()
			}
		}()
	}

	// Retry workers — continuously retry pending IDs.
	for w := 0; w < retryWorkers; w++ {
		go func() {
			for {
				mu.Lock()
				var id int
				found := false
				for k := range pending {
					id = k
					found = true
					break
				}
				mu.Unlock()

				if !found {
					time.Sleep(frontierScannerTickInterval)
					continue
				}

				ok := tryFetch(id)

				mu.Lock()
				if ok {
					if id > highWater {
						highWater = id
					}
					delete(pending, id)
				} else {
					e := pending[id]
					e.retries++
					if e.retries >= maxRetries {
						delete(pending, id)
					} else {
						pending[id] = e
					}
				}
				mu.Unlock()
			}
		}()
	}

	frontierWg.Wait()
}
