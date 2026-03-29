package common

import (
	"sync"
	"time"

	"github.com/plsegott/idstream/seed"
)

type Recorder struct {
	mu     sync.Mutex
	inner  FrontierGetter
	seen   map[int]bool
	result Result
}

func NewRecorder(inner FrontierGetter) *Recorder {
	return &Recorder{
		inner: inner,
		seen:  make(map[int]bool),
		result: Result{
			MinLatency: -1,
		},
	}
}

func (r *Recorder) Get(index int, now time.Time) (seed.Ad, error) {
	ad, err := r.inner.Get(index, now)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.result.Attempts++

	if err != nil {
		if index > r.result.LastIndexSeen {
			r.result.LastIndexSeen = index
		}
		return ad, err
	}

	if index > r.result.LastIndexSeen {
		r.result.LastIndexSeen = index
	}

	if r.seen[index] {
		return ad, nil
	}
	r.seen[index] = true

	latency := now.Sub(ad.LiveAt)
	if latency < 0 {
		latency = 0
	}

	r.result.DiscoveredAds++

	r.result.TotalLatency += latency
	if latency > r.result.MaxLatency {
		r.result.MaxLatency = latency
	}
	if r.result.MinLatency == -1 || latency < r.result.MinLatency {
		r.result.MinLatency = latency
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

func (r *Recorder) GetLatest(now time.Time) (seed.Ad, error) {
	return r.inner.GetLatest(now)
}

func (r *Recorder) RecordAbandoned(index int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.result.AbandonedIDs++
	if index > r.result.LastIndexSeen {
		r.result.LastIndexSeen = index
	}
}

func (r *Recorder) Result() Result {
	out := r.result

	if out.DiscoveredAds > 0 {
		out.AverageLatency = out.TotalLatency / time.Duration(out.DiscoveredAds)
	} else {
		out.MinLatency = 0
	}

	return out
}
