package algorithms

import "time"

const naiveRetryDelay = 1 * time.Second

func Naive(startID int, fetch FetchFunc) {
	id := startID

	for {
		err := fetch(id)
		if err != nil {
			time.Sleep(naiveRetryDelay)
			continue
		}
		id++
	}
}
