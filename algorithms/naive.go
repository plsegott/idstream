package algorithms

import (
	"time"

	"github.com/plsegott/idstream/testing/common"
)

func Naive(getter common.Getter, start time.Time, maxSimTime time.Duration) {
	currentTime := start
	endTime := start.Add(maxSimTime)
	index := 0

	for !currentTime.After(endTime) {
		_, err := getter.GetAd(index, currentTime)

		if err != nil {
			currentTime = currentTime.Add(1 * time.Second)
			continue
		}

		index++
	}
}
