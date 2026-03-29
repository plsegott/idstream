package algorithms

import (
	"strconv"
	"time"
)

const lookaheadTickInterval = 1 * time.Second
const lookaheadDefaultSteps = 500

// Lookahead resolves the live frontier on each tick and sweeps a fixed window
// of `steps` indexes ahead of it in a single thread, minimising missed
// resources by repeatedly covering the active processing zone.
func Lookahead[T IDer](getter FrontierGetter[T], start time.Time, steps int, maxSimTime time.Duration) {
	if steps <= 0 {
		steps = lookaheadDefaultSteps
	}

	currentTime := start
	endTime := start.Add(maxSimTime)
	baseID := -1

	for !currentTime.After(endTime) {
		frontierRes, err := getter.GetLatest(currentTime)
		if err != nil {
			currentTime = currentTime.Add(lookaheadTickInterval)
			continue
		}

		if baseID == -1 {
			first, err := getter.Get(0, currentTime)
			if err != nil {
				currentTime = currentTime.Add(lookaheadTickInterval)
				continue
			}
			id, _ := strconv.Atoi(first.GetID())
			baseID = id
		}

		frontierID, _ := strconv.Atoi(frontierRes.GetID())
		frontierIndex := frontierID - baseID

		for i := frontierIndex; i <= frontierIndex+steps; i++ {
			getter.Get(i, currentTime)
		}

		currentTime = currentTime.Add(lookaheadTickInterval)
	}
}
