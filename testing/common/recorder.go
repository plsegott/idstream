package common

import (
	"time"

	"github.com/plsegott/idstream/seed"
)

type frontierInner interface {
	Getter
	GetLatestLiveAd(now time.Time) (seed.Ad, error)
}

type Recorder struct {
	inner  Getter
	result Result
}

func NewRecorder(inner Getter) *Recorder {
	return &Recorder{
		inner: inner,
		result: Result{
			MinLag: -1,
		},
	}
}

func (r *Recorder) GetAd(index int, now time.Time) (seed.Ad, error) {
	ad, err := r.inner.GetAd(index, now)

	r.result.Attempts++

	if err != nil {
		if index > r.result.LastIndexSeen {
			r.result.LastIndexSeen = index
		}
		return ad, err
	}

	lag := now.Sub(ad.LiveAt)
	if lag < 0 {
		lag = 0
	}

	r.result.DiscoveredAds++
	if ad.Success {
		r.result.SuccessfulAds++
	}

	r.result.TotalLag += lag
	if lag > r.result.MaxLag {
		r.result.MaxLag = lag
	}
	if r.result.MinLag == -1 || lag < r.result.MinLag {
		r.result.MinLag = lag
	}
	if index > r.result.LastIndexSeen {
		r.result.LastIndexSeen = index
	}

	r.result.Discoveries = append(r.result.Discoveries, Discovery{
		Index:        index,
		ID:           ad.Id,
		LiveAt:       ad.LiveAt,
		DiscoveredAt: now,
		Success:      ad.Success,
	})

	return ad, nil
}

func (r *Recorder) GetLatestLiveAd(now time.Time) (seed.Ad, error) {
	if fi, ok := r.inner.(frontierInner); ok {
		return fi.GetLatestLiveAd(now)
	}
	return seed.Ad{}, seed.ErrUnavailable
}

func (r *Recorder) RecordAbandoned(index int) {
	r.result.AbandonedIDs++
	if index > r.result.LastIndexSeen {
		r.result.LastIndexSeen = index
	}
}

func (r *Recorder) Result() Result {
	out := r.result

	if out.DiscoveredAds > 0 {
		out.AverageLag = out.TotalLag / time.Duration(out.DiscoveredAds)
	} else {
		out.MinLag = 0
	}

	return out
}
