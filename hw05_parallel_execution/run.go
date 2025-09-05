package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 || len(tasks) == 0 {
		return nil
	}

	ch := make(chan Task)
	errCh := make(chan struct{})
	var wg sync.WaitGroup
	var errCount int64

	startWorkers(n, ch, errCh, m, &wg)

Loop:
	for _, task := range tasks {
		select {
		case ch <- task:
		case <-errCh:
			if atomic.AddInt64(&errCount, 1) >= int64(m) && m > 0 {
				break Loop
			}
		}
	}
	close(ch)

	wg.Wait()

	if m > 0 && atomic.LoadInt64(&errCount) >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func startWorkers(n int, ch chan Task, errCh chan struct{}, m int, wg *sync.WaitGroup) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range ch {
				if err := task(); err != nil && m > 0 {
					errCh <- struct{}{}
				}
			}
		}()
	}
}
