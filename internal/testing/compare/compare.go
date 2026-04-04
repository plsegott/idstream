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

func RunFrontierScanner(accessor *seed.Accessor, start time.Time, maxWorkers int, maxRetries int, windowSize int, maxRequestsPerSec int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	currentTime := start
	endTime := start.Add(maxSimTime)
	ads := accessor.Ads

	frontierWorkers := maxWorkers / 2
	if frontierWorkers < 1 {
		frontierWorkers = 1
	}

	type pendingEntry struct{ retries int }
	pending := map[int]pendingEntry{}
	highWater := -1

	for !currentTime.After(endTime) {
		createdFrontier := 0
		for i := len(ads) - 1; i >= 0; i-- {
			if !ads[i].CreatedAt.After(currentTime) {
				createdFrontier = i
				break
			}
		}

		scanEnd := highWater + windowSize
		if scanEnd > createdFrontier {
			scanEnd = createdFrontier
		}

		// Frontier scan list — new IDs only.
		frontierScan := make([]int, 0, windowSize)
		for i := highWater + 1; i <= scanEnd; i++ {
			if _, ok := pending[i]; !ok {
				frontierScan = append(frontierScan, i)
			}
		}
		if maxRequestsPerSec > 0 && len(frontierScan) > frontierWorkers {
			frontierScan = frontierScan[:frontierWorkers]
		}

		// Retry scan list — pending IDs only, remaining capacity.
		retryCapacity := maxWorkers - len(frontierScan)
		retryScan := make([]int, 0, len(pending))
		for id := range pending {
			retryScan = append(retryScan, id)
		}
		if maxRequestsPerSec > 0 && len(retryScan) > retryCapacity {
			retryScan = retryScan[:retryCapacity]
		}

		toScan := append(frontierScan, retryScan...)

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
					_, err := rec.Get(results[i].id, currentTime)
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
					rec.RecordAbandoned(r.id)
					delete(pending, r.id)
				} else {
					pending[r.id] = e
				}
			}
		}

		currentTime = currentTime.Add(1 * time.Second)
	}

	return NamedResult{Name: "frontierscanner", Result: rec.Result()}
}

func RunLookahead(accessor *seed.Accessor, start time.Time, steps int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	currentTime := start
	endTime := start.Add(maxSimTime)
	id := 0

	for !currentTime.After(endTime) {
		for i := id; i <= id+steps; i++ {
			_, err := rec.Get(i, currentTime)
			if err == nil {
				id = i + 1
			}
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
