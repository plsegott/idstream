package compare

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/plsegott/idstream/internal/seed"
	"github.com/plsegott/idstream/internal/testing/common"
)

type NamedResult struct {
	Name   string
	Result common.Result
}

func RunNaive(accessor *seed.Accessor, start time.Time, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	currentTime := start
	endTime := start.Add(maxSimTime)
	index := 0

	for !currentTime.After(endTime) {
		_, err := rec.Get(index, currentTime)
		if err != nil {
			currentTime = currentTime.Add(1 * time.Second)
			continue
		}
		index++
	}

	return NamedResult{Name: "naive", Result: rec.Result()}
}

func RunChaser(accessor *seed.Accessor, start time.Time, maxWorkers int, maxAttempts int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)

	type coord struct {
		mu        sync.Mutex
		nextIndex int
	}
	c := &coord{}
	nextIdx := func() int {
		c.mu.Lock()
		defer c.mu.Unlock()
		i := c.nextIndex
		c.nextIndex++
		return i
	}

	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			currentTime := start
			endTime := start.Add(maxSimTime)
			index := nextIdx()
			attempts := 0

			for !currentTime.After(endTime) {
				_, err := rec.Get(index, currentTime)
				if err != nil {
					attempts++
					if attempts > maxAttempts {
						rec.RecordAbandoned(index)
						index = nextIdx()
						attempts = 0
						continue
					}
					currentTime = currentTime.Add(10 * time.Second)
					continue
				}
				index = nextIdx()
				attempts = 0
			}
		}()
	}
	wg.Wait()

	return NamedResult{Name: "chaser", Result: rec.Result()}
}

func RunBinaryFrontier(accessor *seed.Accessor, start time.Time, maxWorkers int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	currentTime := start
	endTime := start.Add(maxSimTime)
	nextIndex := 0

	type pendingEntry struct{ firstAttempt time.Time }
	pending := map[int]pendingEntry{}
	maxDelay := 8 * time.Minute

	for !currentTime.After(endTime) {
		frontierIndex, err := rec.GetLatestIndex(currentTime)
		if err != nil {
			currentTime = currentTime.Add(1 * time.Second)
			continue
		}

		toScan := make([]int, 0, len(pending)+(frontierIndex-nextIndex+1))
		for idx, entry := range pending {
			if currentTime.Sub(entry.firstAttempt) < maxDelay {
				toScan = append(toScan, idx)
			} else {
				delete(pending, idx)
			}
		}
		for i := nextIndex; i <= frontierIndex; i++ {
			if _, ok := pending[i]; !ok {
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
					_, err := rec.Get(results[i].index, currentTime)
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
		currentTime = currentTime.Add(10 * time.Second)
	}

	return NamedResult{Name: "binaryfrontier", Result: rec.Result()}
}

func RunLookahead(accessor *seed.Accessor, start time.Time, steps int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	currentTime := start
	endTime := start.Add(maxSimTime)

	for !currentTime.After(endTime) {
		frontierIndex, err := rec.GetLatestIndex(currentTime)
		if err != nil {
			currentTime = currentTime.Add(1 * time.Second)
			continue
		}

		for i := frontierIndex; i <= frontierIndex+steps; i++ {
			rec.Get(i, currentTime)
		}

		currentTime = currentTime.Add(1 * time.Second)
	}

	return NamedResult{Name: "lookahead", Result: rec.Result()}
}

func PrintResults(results []NamedResult, totalSuccessful int) {
	fmt.Printf("Total discoverable resources: %d\n\n", totalSuccessful)
	fmt.Printf(
		"%-16s %-10s %-12s %-10s %-12s %-12s %-12s\n",
		"Algorithm", "Attempts", "Discovered", "Coverage", "Abandoned", "AvgLatency", "MaxLatency",
	)

	for _, r := range results {
		coverage := "0.0%"
		if totalSuccessful > 0 {
			coverage = fmt.Sprintf("%.1f%%", float64(r.Result.DiscoveredAds)/float64(totalSuccessful)*100)
		}
		fmt.Printf(
			"%-16s %-10d %-12d %-10s %-12d %-12s %-12s\n",
			r.Name,
			r.Result.Attempts,
			r.Result.DiscoveredAds,
			coverage,
			r.Result.AbandonedIDs,
			r.Result.AverageLatency.String(),
			r.Result.MaxLatency.String(),
		)
	}
}
