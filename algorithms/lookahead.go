package algorithms

import (
	"strconv"
	"time"

	"github.com/plsegott/idstream/testing/common"
)

const lookaheadTickInterval = 1 * time.Second
const lookaheadDefaultSteps = 500

// Lookahead resolves the live frontier on each tick and sweeps a fixed window
// of `steps` indexes ahead of it in a single thread, minimising missed
// resources by repeatedly covering the active processing zone.
func Lookahead(getter common.Getter, start time.Time, steps int, maxSimTime time.Duration) {
	fg, ok := getter.(frontierGetter)
	if !ok {
		Naive(getter, start, maxSimTime)
		return
	}

	if steps <= 0 {
		steps = lookaheadDefaultSteps
	}

	currentTime := start
	endTime := start.Add(maxSimTime)
	baseID := -1

	for !currentTime.After(endTime) {
		frontierAd, err := fg.GetLatestLiveAd(currentTime)
		if err != nil {
			currentTime = currentTime.Add(lookaheadTickInterval)
			continue
		}

		if baseID == -1 {
			first, err := fg.GetAd(0, currentTime)
			if err != nil {
				currentTime = currentTime.Add(lookaheadTickInterval)
				continue
			}
			id, _ := strconv.Atoi(first.Id)
			baseID = id
		}

		frontierID, _ := strconv.Atoi(frontierAd.Id)
		frontierIndex := frontierID - baseID

		for i := frontierIndex; i <= frontierIndex+steps; i++ {
			fg.GetAd(i, currentTime)
		}

		currentTime = currentTime.Add(lookaheadTickInterval)
	}
}
