package algorithms

import (
	"sync"
	"time"

	"github.com/plsegott/idstream/testing/common"
)

type coordinator struct {
	mu        sync.Mutex
	nextIndex int
}

func (c *coordinator) GetNextIndex() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	i := c.nextIndex
	c.nextIndex++
	return i
}

type abandonRecorder interface {
	RecordAbandoned(index int)
}

func Chaser(getter common.Getter, start time.Time, maxWorkers int, maxAttempts int, maxSimTime time.Duration) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	coord := &coordinator{}
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(getter, start, coord, maxAttempts, maxSimTime)
		}()
	}

	wg.Wait()
}

func runWorker(getter common.Getter, start time.Time, coord *coordinator, maxAttempts int, maxSimTime time.Duration) {
	currentTime := start
	endTime := start.Add(maxSimTime)

	index := coord.GetNextIndex()
	attemptsForCurrentIndex := 0

	resetWorker := func() {
		index = coord.GetNextIndex()
		attemptsForCurrentIndex = 0
	}

	for !currentTime.After(endTime) {
		_, err := getter.GetAd(index, currentTime)

		if err != nil {
			attemptsForCurrentIndex++

			if attemptsForCurrentIndex > maxAttempts {
				if ar, ok := getter.(abandonRecorder); ok {
					ar.RecordAbandoned(index)
				}
				resetWorker()
				continue
			}

			currentTime = currentTime.Add(10 * time.Second)
			continue
		}

		resetWorker()
	}
}
