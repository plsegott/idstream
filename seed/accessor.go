package seed

import "time"

type Accessor struct {
	Ads []Ad
}

func NewAccessor(ads []Ad) *Accessor {
	return &Accessor{Ads: ads}
}

// CountSuccessful returns the total number of ads that will eventually go live
// (Success=true), regardless of when they become available.
func CountSuccessful(ads []Ad) int {
	n := 0
	for _, ad := range ads {
		if ad.Success {
			n++
		}
	}
	return n
}

func (a *Accessor) GetLatestLiveAd(now time.Time) (Ad, error) {
	for i := len(a.Ads) - 1; i >= 0; i-- {
		ad := a.Ads[i]
		if ad.Success && !now.Before(ad.LiveAt) {
			return ad, nil
		}
	}
	return Ad{}, ErrUnavailable
}

func (a *Accessor) GetAd(index int, now time.Time) (Ad, error) {
	if index < 0 || index >= len(a.Ads) {
		return Ad{}, ErrUnavailable
	}

	ad := a.Ads[index]

	if now.Before(ad.CreatedAt) {
		return Ad{}, ErrUnavailable
	}

	if !ad.Success {
		return Ad{}, ErrUnavailable
	}

	if now.Before(ad.LiveAt) {
		return Ad{}, ErrUnavailable
	}

	return ad, nil
}
