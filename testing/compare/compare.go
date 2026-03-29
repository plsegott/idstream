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
	algorithms.Naive(rec, start, maxSimTime)
	return NamedResult{
		Name:   "naive",
		Result: rec.Result(),
	}
}

func RunChaser(accessor *seed.Accessor, start time.Time, maxWorkers int, maxAttempts int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.Chaser(rec, start, maxWorkers, maxAttempts, maxSimTime)
	return NamedResult{
		Name:   "chaser",
		Result: rec.Result(),
	}
}

func RunBinaryFrontier(accessor *seed.Accessor, start time.Time, maxWorkers int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.BinaryFrontier(rec, start, maxWorkers, maxSimTime)
	return NamedResult{
		Name:   "binaryfrontier",
		Result: rec.Result(),
	}
}

func RunLookahead(accessor *seed.Accessor, start time.Time, steps int, maxSimTime time.Duration) NamedResult {
	rec := common.NewRecorder(accessor)
	algorithms.Lookahead(rec, start, steps, maxSimTime)
	return NamedResult{
		Name:   "lookahead",
		Result: rec.Result(),
	}
}

func PrintResults(results []NamedResult) {
	fmt.Printf(
		"%-12s %-10s %-12s %-12s %-12s %-12s %-12s %-12s\n",
		"Algorithm", "Attempts", "Discovered", "Successful", "Abandoned", "AvgLag", "MaxLag", "LastIndex",
	)

	for _, r := range results {
		fmt.Printf(
			"%-12s %-10d %-12d %-12d %-12d %-12s %-12s %-12d\n",
			r.Name,
			r.Result.Attempts,
			r.Result.DiscoveredAds,
			r.Result.SuccessfulAds,
			r.Result.AbandonedIDs,
			r.Result.AverageLag.String(),
			r.Result.MaxLag.String(),
			r.Result.LastIndexSeen,
		)
	}
}
