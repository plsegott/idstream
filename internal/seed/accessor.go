package seed

import "time"

type Accessor struct {
	Ads []Ad
}

func NewAccessor(ads []Ad) *Accessor {
	return &Accessor{Ads: ads}
}

func CountSuccessful(ads []Ad) int {
	n := 0
	for _, ad := range ads {
		if ad.Success {
			n++
		}
	}
	return n
}

func (a *Accessor) GetLatestIndex(now time.Time) (int, error) {
	for i := len(a.Ads) - 1; i >= 0; i-- {
		ad := a.Ads[i]
		if ad.Success && !now.Before(ad.LiveAt) {
			return i, nil
		}
	}
	return 0, ErrUnavailable
}

func (a *Accessor) Get(index int, now time.Time) (Ad, error) {
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
