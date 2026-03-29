package algorithms

import "time"

// Getter is the core interface for sequential ID discovery.
// Implement this against any API that assigns monotonically increasing integer IDs.
type Getter[T any] interface {
	Get(index int, now time.Time) (T, error)
}

// FrontierGetter extends Getter with the ability to resolve the index of the
// current live frontier — the highest index that is available right now.
// Required by BinaryFrontier and Lookahead.
type FrontierGetter[T any] interface {
	Getter[T]
	GetLatestIndex(now time.Time) (int, error)
}
