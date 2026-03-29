package eval

import (
	"time"

	"github.com/plsegott/idstream/seed"
	"github.com/plsegott/idstream/testing/common"
)

type Stats struct {
	TotalAds             int
	DiscoveredAds        int
	MissedAds            int
	CoverageRate         float64
	AverageDiscoveredLag time.Duration
	MaxDiscoveredLag     time.Duration

	AverageFrontierGap float64
	MaxFrontierGap     int
}

func Evaluate(ads []seed.Ad, result common.Result) Stats {
	stats := Stats{
		TotalAds: len(ads),
	}

	discoveredByIndex := make(map[int]common.Discovery, len(result.Discoveries))
	for _, d := range result.Discoveries {
		discoveredByIndex[d.Index] = d
	}

	stats.DiscoveredAds = len(result.Discoveries)
	stats.MissedAds = len(ads) - stats.DiscoveredAds

	if len(ads) > 0 {
		stats.CoverageRate = float64(stats.DiscoveredAds) / float64(len(ads))
	}

	stats.AverageDiscoveredLag = result.AverageLag
	stats.MaxDiscoveredLag = result.MaxLag

	var totalGap int
	var gapSamples int

	for i, ad := range ads {
		_ = ad

		d, ok := discoveredByIndex[i]
		if !ok {
			continue
		}

		frontierAtDiscovery := highestLiveIndexAt(ads, d.DiscoveredAt)
		gap := frontierAtDiscovery - i
		if gap < 0 {
			gap = 0
		}

		totalGap += gap
		gapSamples++

		if gap > stats.MaxFrontierGap {
			stats.MaxFrontierGap = gap
		}
	}

	if gapSamples > 0 {
		stats.AverageFrontierGap = float64(totalGap) / float64(gapSamples)
	}

	return stats
}

func highestLiveIndexAt(ads []seed.Ad, t time.Time) int {
	highest := -1
	for i, ad := range ads {
		if !ad.LiveAt.After(t) {
			highest = i
		}
	}
	return highest
}
