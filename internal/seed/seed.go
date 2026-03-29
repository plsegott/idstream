package seed

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Ad struct {
	Id        string
	CreatedAt time.Time
	LiveAt    time.Time
	Success   bool
}


type Seed struct {
	Ads    []Ad
	Rng    *rand.Rand
	NextID int64
}

type Phase struct {
	Name         string
	Duration     time.Duration
	AvgAdsPerSec float64
}

func CreateSeed(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

const minProcessingDelay = 2 * time.Minute
const maxProcessingDelayRange = 5 * time.Minute

func GenerateAd(r *rand.Rand, created time.Time, id int64) Ad {
	delay := minProcessingDelay + time.Duration(r.Int63n(int64(maxProcessingDelayRange)))

	return Ad{
		Id:        fmt.Sprintf("%d", id),
		CreatedAt: created,
		LiveAt:    created.Add(delay),
		Success:   r.Intn(2) == 1,
	}
}

// poisson returns a Poisson-distributed random integer with mean lambda,
// using the Knuth algorithm.
func poisson(r *rand.Rand, lambda float64) int {
	l := math.Exp(-lambda)
	k := 0
	p := 1.0

	for p > l {
		k++
		p *= r.Float64()
	}

	return k - 1
}

func RunSeedProfile(phases []Phase) *Seed {
	s := &Seed{
		Ads:    []Ad{},
		Rng:    CreateSeed(1),
		NextID: 1000,
	}

	currentTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, phase := range phases {
		seconds := int(phase.Duration.Seconds())

		for i := 0; i < seconds; i++ {
			count := poisson(s.Rng, phase.AvgAdsPerSec)

			for j := 0; j < count; j++ {
				created := currentTime.Add(time.Duration(j) * time.Millisecond)

				ad := GenerateAd(s.Rng, created, s.NextID)
				s.Ads = append(s.Ads, ad)
				s.NextID++
			}

			currentTime = currentTime.Add(1 * time.Second)
		}
	}

	return s
}
