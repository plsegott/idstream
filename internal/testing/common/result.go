package common

import "time"

type Discovery struct {
	Index        int
	ID           string
	LiveAt       time.Time
	DiscoveredAt time.Time
	Success      bool
}

type Result struct {
	Name string

	Attempts      int
	DiscoveredAds int
	SuccessfulAds int
	AbandonedIDs  int
	LastIndexSeen int

	TotalLatency   time.Duration
	MaxLatency     time.Duration
	MinLatency     time.Duration
	AverageLatency time.Duration

	Discoveries []Discovery
}
