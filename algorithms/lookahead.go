package algorithms

import "time"

const lookaheadTickInterval = 1 * time.Second
const lookaheadDefaultSteps = 500

// Lookahead scans a fixed window of `steps` IDs ahead of the current position
// on every tick, in a single thread.
func Lookahead(startID int, steps int, fetch FetchFunc) {
	if steps <= 0 {
		steps = lookaheadDefaultSteps
	}

	id := startID

	for {
		for i := id; i <= id+steps; i++ {
			if fetch(i) == nil {
				id = i + 1
			}
		}
		time.Sleep(lookaheadTickInterval)
	}
}
