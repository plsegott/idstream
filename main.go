package main

import (
	"time"

	"github.com/plsegott/idstream/seed"
	"github.com/plsegott/idstream/testing/compare"
)

func main() {
	phases := []seed.Phase{
		{Name: "slow", Duration: 8 * time.Hour, AvgAdsPerSec: 0.8},
		{Name: "burst", Duration: 8 * time.Hour, AvgAdsPerSec: 3.5},
		{Name: "cooldown", Duration: 8 * time.Hour, AvgAdsPerSec: 1.2},
	}

	s := seed.RunSeedProfile(phases)
	accessor := seed.NewAccessor(s.Ads)
	start := s.Ads[0].CreatedAt

	results := []compare.NamedResult{
		compare.RunNaive(accessor, start, 24*time.Hour),
		compare.RunChaser(accessor, start, 200, 5, 24*time.Hour),
		compare.RunBinaryFrontier(accessor, start, 200, 24*time.Hour),
		compare.RunLookahead(accessor, start, 500, 24*time.Hour),
	}

	compare.PrintResults(results, seed.CountSuccessful(s.Ads))
}
