package algorithms

import "time"

// Getter is the core interface for sequential ID discovery.
// Implement this against any API that assigns monotonically increasing integer IDs.
type Getter[T any] interface {
	Get(index int, now time.Time) (T, error)
}

// FrontierGetter extends Getter with the ability to resolve the current live
// frontier — the latest resource that is available right now. Algorithms that
// implement frontier-aware strategies (BinaryFrontier, Lookahead) require this.
type FrontierGetter[T any] interface {
	Getter[T]
	GetLatest(now time.Time) (T, error)
}
