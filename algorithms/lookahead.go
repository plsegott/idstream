package algorithms

import (
	"strconv"
	"time"

	"github.com/plsegott/idstream/testing/common"
)

// Lookahead finds the current live frontier and scans the next `steps` indexes
// ahead of it on every tick, in a single thread.
func Lookahead(getter common.Getter, start time.Time, steps int, maxSimTime time.Duration) {
	fg, ok := getter.(frontierGetter)
	if !ok {
		Naive(getter, start, maxSimTime)
		return
	}

	if steps <= 0 {
		steps = 500
	}

	currentTime := start
	endTime := start.Add(maxSimTime)
	baseID := -1

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

		for i := frontierIndex; i <= frontierIndex+steps; i++ {
			fg.GetAd(i, currentTime)
		}

		currentTime = currentTime.Add(1 * time.Second)
	}
}
