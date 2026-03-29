package algorithms

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// maxProcessingDelay is the upper bound on resource processing time
// (seed generates 2–7 minute delays). After this window a persistent error
// indicates permanent failure and the index is evicted from the retry set.
const maxProcessingDelay = 8 * time.Minute
const binaryFrontierTickInterval = 10 * time.Second
const binaryFrontierErrorBackoff = 1 * time.Second

type pendingEntry struct {
	firstAttempt time.Time
}

// IDer is an optional interface resources can implement to expose their
// sequential integer ID. BinaryFrontier uses this to derive slice indexes
// arithmetically from the frontier resource.
type IDer interface {
	GetID() string
}

// BinaryFrontier discovers resources by:
//  1. Calling GetLatest to anchor the live frontier.
//  2. Deriving the frontier's slice index arithmetically from the sequential ID
//     (frontierIndex = frontierID - baseID).
//  3. Fanning out a concurrent scan over [nextIndex, frontierIndex].
//  4. Maintaining a retry set for indexes not yet live at scan time —
//     retried each round until maxProcessingDelay has elapsed, after which
//     persistent errors are treated as permanent failures.
func BinaryFrontier[T IDer](getter FrontierGetter[T], start time.Time, maxWorkers int, maxSimTime time.Duration) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	currentTime := start
	endTime := start.Add(maxSimTime)
	nextIndex := 0
	baseID := -1

	// pending holds indexes that returned an error but may still go live.
	pending := map[int]pendingEntry{}

	for !currentTime.After(endTime) {
		frontierRes, err := getter.GetLatest(currentTime)
		if err != nil {
			currentTime = currentTime.Add(binaryFrontierErrorBackoff)
			continue
		}

		if baseID == -1 {
			first, err := getter.Get(0, currentTime)
			if err != nil {
				currentTime = currentTime.Add(binaryFrontierErrorBackoff)
				continue
			}
			id, _ := strconv.Atoi(first.GetID())
			baseID = id
		}

		frontierID, _ := strconv.Atoi(frontierRes.GetID())
		frontierIndex := frontierID - baseID

		// Build the scan list: pending retries + new indexes up to the frontier.
		toScan := make([]int, 0, len(pending)+(frontierIndex-nextIndex+1))

		for idx, entry := range pending {
			if currentTime.Sub(entry.firstAttempt) < maxProcessingDelay {
				toScan = append(toScan, idx)
			} else {
				delete(pending, idx)
			}
		}

		for i := nextIndex; i <= frontierIndex; i++ {
			if _, alreadyPending := pending[i]; !alreadyPending {
				toScan = append(toScan, i)
			}
		}

		type scanResult struct {
			index int
			ok    bool
		}
		results := make([]scanResult, len(toScan))
		for i, idx := range toScan {
			results[i].index = idx
		}

		var cursor atomic.Int64
		cursor.Store(0)

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
					_, err := getter.Get(results[i].index, currentTime)
					results[i].ok = err == nil
				}
			}()
		}
		wg.Wait()

		for _, r := range results {
			if r.ok {
				delete(pending, r.index)
			} else if r.index >= nextIndex {
				if _, exists := pending[r.index]; !exists {
					pending[r.index] = pendingEntry{firstAttempt: currentTime}
				}
			}
		}

		nextIndex = frontierIndex + 1
		currentTime = currentTime.Add(binaryFrontierTickInterval)
	}
}
