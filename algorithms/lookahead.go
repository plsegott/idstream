package algorithms

import "time"

const lookaheadTickInterval = 1 * time.Second
const lookaheadDefaultSteps = 500

// Lookahead resolves the live frontier on each tick and sweeps a fixed window
// of `steps` IDs ahead of it, minimising missed resources by repeatedly
// covering the active processing zone.
func Lookahead(startID int, steps int, fetch FetchFunc, latest LatestFunc) {
	if steps <= 0 {
		steps = lookaheadDefaultSteps
	}

	for {
		frontierID, err := latest()
		if err != nil {
			time.Sleep(lookaheadTickInterval)
			continue
		}

		for id := frontierID; id <= frontierID+steps; id++ {
			fetch(id)
		}

		time.Sleep(lookaheadTickInterval)
	}
}
