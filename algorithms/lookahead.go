package algorithms

import "time"

const lookaheadTickInterval = 1 * time.Second
const lookaheadDefaultSteps = 500

// Lookahead resolves the live frontier on each tick and sweeps a fixed window
// of `steps` indexes ahead of it in a single thread, minimising missed
// resources by repeatedly covering the active processing zone.
func Lookahead[T any](getter FrontierGetter[T], start time.Time, steps int, maxSimTime time.Duration) {
	if steps <= 0 {
		steps = lookaheadDefaultSteps
	}

	currentTime := start
	endTime := start.Add(maxSimTime)

	for !currentTime.After(endTime) {
		frontierIndex, err := getter.GetLatestIndex(currentTime)
		if err != nil {
			currentTime = currentTime.Add(lookaheadTickInterval)
			continue
		}

		for i := frontierIndex; i <= frontierIndex+steps; i++ {
			getter.Get(i, currentTime)
		}

		currentTime = currentTime.Add(lookaheadTickInterval)
	}
}
