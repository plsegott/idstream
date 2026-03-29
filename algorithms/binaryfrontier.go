package algorithms

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/plsegott/idstream/seed"
	"github.com/plsegott/idstream/testing/common"
)

type frontierGetter interface {
	common.Getter
	GetLatestLiveAd(now time.Time) (seed.Ad, error)
}

// maxAdDelay is the upper bound on how long an ad can take to go live after
// creation (seed generates 2–7 minute delays). After this window an error
// means the ad permanently failed and we stop retrying.
const maxAdDelay = 8 * time.Minute

type pendingEntry struct {
	firstAttempt time.Time
}

// BinaryFrontier discovers ads by:
//  1. Calling GetLatestLiveAd to anchor the live frontier.
//  2. Computing the frontier slice index from the ID difference (IDs are
//     sequential integers, so frontierIndex = frontierID - baseID).
//  3. Scanning [nextIndex, frontierIndex] concurrently with workers.
//  4. Keeping a retry window for indexes that failed — they get re-attempted
//     each round until maxAdDelay has elapsed (after which they are
//     permanently dead due to Success=false in the seed).
func BinaryFrontier(getter common.Getter, start time.Time, maxWorkers int, maxSimTime time.Duration) {
	fg, ok := getter.(frontierGetter)
	if !ok {
		Naive(getter, start, maxSimTime)
		return
	}

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
		frontierAd, err := fg.GetLatestLiveAd(currentTime)
		if err != nil {
			currentTime = currentTime.Add(1 * time.Second)
			continue
		}

		if baseID == -1 {
			first, err := fg.GetAd(0, currentTime)
			if err != nil {
				currentTime = currentTime.Add(1 * time.Second)
				continue
			}
			id, _ := strconv.Atoi(first.Id)
			baseID = id
		}

		frontierID, _ := strconv.Atoi(frontierAd.Id)
		frontierIndex := frontierID - baseID

		// Build the scan list: pending retries + new indexes up to the frontier.
		toScan := make([]int, 0, len(pending)+(frontierIndex-nextIndex+1))

		for idx, entry := range pending {
			if currentTime.Sub(entry.firstAttempt) < maxAdDelay {
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

		// Scan concurrently, collecting failed indexes for the retry window.
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
					_, err := fg.GetAd(results[i].index, currentTime)
					results[i].ok = err == nil
				}
			}()
		}
		wg.Wait()

		// Update pending based on scan results.
		for _, r := range results {
			if r.ok {
				delete(pending, r.index)
			} else if r.index >= nextIndex {
				// Only add to pending if within the new range (not already pending).
				if _, exists := pending[r.index]; !exists {
					pending[r.index] = pendingEntry{firstAttempt: currentTime}
				}
			}
		}

		nextIndex = frontierIndex + 1
		currentTime = currentTime.Add(10 * time.Second)
	}
}
