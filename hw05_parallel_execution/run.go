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
	done := make(chan struct{})
	var wg sync.WaitGroup
	var errCount int64
	var once sync.Once

	startDispatcher(tasks, ch, done, &wg)

	startWorkers(n, ch, done, m, &errCount, &once, &wg)

	wg.Wait()

	if m > 0 && atomic.LoadInt64(&errCount) >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func startDispatcher(tasks []Task, ch chan Task, done chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, task := range tasks {
			select {
			case <-done:
				return
			case ch <- task:
			}
		}
		close(ch)
	}()
}

func startWorkers(n int, ch chan Task, done chan struct{}, m int, errC *int64, once *sync.Once, wg *sync.WaitGroup) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case task, ok := <-ch:
					if !ok {
						return
					}
					if err := task(); err != nil && m > 0 {
						if atomic.AddInt64(errC, 1) >= int64(m) {
							once.Do(func() { close(done) })
							return
						}
					}
				}
			}
		}()
	}
}
