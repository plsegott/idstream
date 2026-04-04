package algorithms

import (
	"sync"
	"sync/atomic"
	"time"
)

const frontierScannerTickInterval = 1 * time.Second

// FrontierScanner discovers resources by scanning a window ahead of the highest
// known live ID. Failed IDs are retried up to maxRetries times to handle
// resources that are created but not yet live. maxRequestsPerSec limits the
// total number of fetch calls per second across all workers (0 = unlimited).
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

	type pendingEntry struct {
		retries int
	}

	// highWater is the highest ID we've ever seen succeed.
	// We always scan [highWater+1, highWater+windowSize] for new IDs.
	highWater := startID - 1
	pending := map[int]pendingEntry{}

	var limiter <-chan time.Time
	if maxRequestsPerSec > 0 {
		ticker := time.NewTicker(time.Second / time.Duration(maxRequestsPerSec))
		defer ticker.Stop()
		limiter = ticker.C
	}

	for {
		scanStart := highWater + 1
		scanEnd := highWater + windowSize

		// New IDs at the frontier get priority. Pending retries fill remaining capacity.
		toScan := make([]int, 0, windowSize+len(pending))
		for i := scanStart; i <= scanEnd; i++ {
			if _, ok := pending[i]; !ok {
				toScan = append(toScan, i)
			}
		}
		if maxRequestsPerSec <= 0 || len(toScan) < maxRequestsPerSec {
			for id := range pending {
				toScan = append(toScan, id)
			}
		}

		type scanResult struct {
			id int
			ok bool
		}
		results := make([]scanResult, len(toScan))
		for i, id := range toScan {
			results[i].id = id
		}

		var cursor atomic.Int64
		var wg sync.WaitGroup
		for w := 0; w < maxWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					i := int(cursor.Add(1) - 1)
					if i >= len(results) {
						return
					}
					if limiter != nil {
						<-limiter
					}
					err := fetch(results[i].id)
					results[i].ok = err == nil
				}
			}()
		}
		wg.Wait()

		for _, r := range results {
			if r.ok {
				delete(pending, r.id)
				if r.id > highWater {
					highWater = r.id
				}
			} else {
				e := pending[r.id]
				e.retries++
				if e.retries >= maxRetries {
					delete(pending, r.id)
				} else {
					pending[r.id] = e
				}
			}
		}

		time.Sleep(frontierScannerTickInterval)
	}
}
