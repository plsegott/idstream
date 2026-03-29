package compare

import (
	"fmt"
	"time"

	"github.com/plsegott/idstream/algorithms"
	"github.com/plsegott/idstream/seed"
	"github.com/plsegott/idstream/testing/common"
)

type NamedResult struct {
	Name   string
	Result common.Result
}

func RunNaive(accessor *seed.Accessor, start time.Time, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.Naive[seed.Ad](rec, start, maxSimTime)
	return NamedResult{Name: "naive", Result: rec.Result()}
}

func RunChaser(accessor *seed.Accessor, start time.Time, maxWorkers int, maxAttempts int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.Chaser[seed.Ad](rec, start, maxWorkers, maxAttempts, maxSimTime)
	return NamedResult{Name: "chaser", Result: rec.Result()}
}

func RunBinaryFrontier(accessor *seed.Accessor, start time.Time, maxWorkers int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.BinaryFrontier[seed.Ad](rec, start, maxWorkers, maxSimTime)
	return NamedResult{Name: "binaryfrontier", Result: rec.Result()}
}

func RunLookahead(accessor *seed.Accessor, start time.Time, steps int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.Lookahead[seed.Ad](rec, start, steps, maxSimTime)
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
