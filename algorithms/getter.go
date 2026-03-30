package algorithms

// FetchFunc is called for each discovered ID.
// Return nil on success, an error if the resource is unavailable or not yet live.
type FetchFunc func(id int) error
