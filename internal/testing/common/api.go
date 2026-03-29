package common

import (
	"github.com/plsegott/idstream/algorithms"
	"github.com/plsegott/idstream/internal/seed"
)

// Getter and FrontierGetter are re-exported here for convenience so benchmark
// code only needs to import testing/common.
type Getter = algorithms.Getter[seed.Ad]
type FrontierGetter = algorithms.FrontierGetter[seed.Ad]
