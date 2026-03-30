package algorithms

import (
	"sync"
	"sync/atomic"
	"time"
)

const maxProcessingDelay = 8 * time.Minute
const binaryFrontierRetryDelay = 1 * time.Second
const binaryFrontierTickInterval = 10 * time.Second

type pendingEntry struct {
	firstAttempt time.Time
}

// BinaryFrontier discovers resources by:
//  1. Calling latest to anchor the live frontier.
//  2. Fanning out a concurrent scan over [nextID, frontierID].
//  3. Maintaining a retry set for IDs not yet live at scan time —
//     retried each round until maxProcessingDelay has elapsed, after which
//     persistent errors are treated as permanent failures.
func BinaryFrontier(startID int, maxWorkers int, fetch FetchFunc, latest LatestFunc) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	nextID := startID
	pending := map[int]pendingEntry{}

	for {
		frontierID, err := latest()
		if err != nil {
			time.Sleep(binaryFrontierRetryDelay)
			continue
		}

		now := time.Now()
		toScan := make([]int, 0, len(pending)+(frontierID-nextID+1))

		for id, entry := range pending {
			if now.Sub(entry.firstAttempt) < maxProcessingDelay {
				toScan = append(toScan, id)
			} else {
				delete(pending, id)
			}
		}

		for i := nextID; i <= frontierID; i++ {
			if _, alreadyPending := pending[i]; !alreadyPending {
				toScan = append(toScan, i)
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
					err := fetch(results[i].id)
					results[i].ok = err == nil
				}
			}()
		}
		wg.Wait()

		for _, r := range results {
			if r.ok {
				delete(pending, r.id)
			} else if r.id >= nextID {
				if _, exists := pending[r.id]; !exists {
					pending[r.id] = pendingEntry{firstAttempt: now}
				}
			}
		}

		nextID = frontierID + 1
		time.Sleep(binaryFrontierTickInterval)
	}
}
