package main

import "time"

func Batcher[T any](input <-chan T, interval time.Duration, process func([]T)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	batch := make([]T, 0)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		process(batch)
		batch = make([]T, 0)
	}

	for {
		select {
		case item, ok := <-input:
			if !ok {
				flush()
				return
			}

			batch = append(batch, item)

		case <-ticker.C:
			flush()
		}
	}
}
