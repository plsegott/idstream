package main

import (
	"time"

	"github.com/plsegott/idstream/internal/seed"
	"github.com/plsegott/idstream/internal/testing/compare"
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
		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 100, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 150, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 200, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 250, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 300, 24*time.Hour),

		compare.RunFrontierScanner(accessor, start, 200, 120, 800, 150, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 240, 800, 150, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 360, 800, 150, 24*time.Hour),
		compare.RunFrontierScanner(accessor, start, 200, 480, 800, 150, 24*time.Hour),
	}

	compare.PrintResults(results, seed.CountSuccessful(s.Ads))
}
