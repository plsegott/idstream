package main

import (
	"time"

	"github.com/plsegott/idstream/internal/seed"
	"github.com/plsegott/idstream/internal/testing/compare"
)

func main() {
	phases := []seed.Phase{
		{Name: "slow", Duration: 8 * time.Hour, AvgAdsPerSec: 10},
		{Name: "burst", Duration: 8 * time.Hour, AvgAdsPerSec: 60},
		{Name: "cooldown", Duration: 8 * time.Hour, AvgAdsPerSec: 30},
	}

	s := seed.RunSeedProfile(phases)
	accessor := seed.NewAccessor(s.Ads)
	start := s.Ads[0].CreatedAt

	results := []compare.NamedResult{
		compare.RunFrontierScanner(accessor, start, 200, 240, 400, 50, 24*time.Hour),
	}

	compare.PrintResults(results, seed.CountSuccessful(s.Ads))
}
