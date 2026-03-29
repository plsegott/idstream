package algorithms

import (
	"time"

	"github.com/plsegott/idstream/testing/common"
)

const naiveTickInterval = 1 * time.Second

func Naive(getter common.Getter, start time.Time, maxSimTime time.Duration) {
	currentTime := start
	endTime := start.Add(maxSimTime)
	index := 0

	for !currentTime.After(endTime) {
		_, err := getter.GetAd(index, currentTime)

		if err != nil {
			currentTime = currentTime.Add(naiveTickInterval)
			continue
		}

		index++
	}
}
