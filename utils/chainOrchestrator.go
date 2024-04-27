package utils

import "sync"

func ChainOrchestrator[R, T any](ch <-chan T, fn func(T) R) chan R {

	wg := sync.WaitGroup{}
	out := make(chan R)

	action := func(v T) {
		defer wg.Done()
		vl := fn(v)
		out <- vl
	}

	wg.Add(1)
	go func() {
		for vl := range ch {
			wg.Add(1)
			go action(vl)
		}
		wg.Done()
	}()
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
