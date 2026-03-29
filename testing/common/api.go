package common

import (
	"time"

	"github.com/plsegott/idstream/seed"
)

type Getter interface {
	GetAd(index int, now time.Time) (seed.Ad, error)
}

type FrontierGetter interface {
	Getter
	GetLatestLiveAd(now time.Time) (seed.Ad, error)
}
