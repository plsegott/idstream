package algorithms

import "time"

const naiveTickInterval = 1 * time.Second

func Naive[T any](getter Getter[T], start time.Time, maxSimTime time.Duration) {
	currentTime := start
	endTime := start.Add(maxSimTime)
	index := 0

	for !currentTime.After(endTime) {
		_, err := getter.Get(index, currentTime)

		if err != nil {
			currentTime = currentTime.Add(naiveTickInterval)
			continue
		}

		index++
	}
}
