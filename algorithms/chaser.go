package algorithms

import (
	"sync"
	"time"
)

const chaserRetryDelay = 10 * time.Second

type coordinator struct {
	mu        sync.Mutex
	nextID    int
}

func (c *coordinator) Next() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextID
	c.nextID++
	return id
}

func Chaser(startID int, maxWorkers int, maxAttempts int, fetch FetchFunc) {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	coord := &coordinator{nextID: startID}
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := coord.Next()
			attempts := 0

			for {
				err := fetch(id)
				if err != nil {
					attempts++
					if attempts > maxAttempts {
						id = coord.Next()
						attempts = 0
						continue
					}
					time.Sleep(chaserRetryDelay)
					continue
				}
				id = coord.Next()
				attempts = 0
			}
		}()
	}

	wg.Wait()
}
