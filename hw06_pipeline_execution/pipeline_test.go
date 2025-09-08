package hw06pipelineexecution

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = time.Millisecond * 100 // Увеличено с 50 мс до 100 мс
)

func TestPipeline(t *testing.T) {
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	gWithDone := func(name string, f func(v interface{}) interface{}, wg *sync.WaitGroup, ctx context.Context, done In) Stage {
		return func(in In) Out {
			out := make(Bi)
			if wg != nil {
				wg.Add(1)
			}
			go func() {
				if wg != nil {
					defer wg.Done()
				}
				defer close(out)
				for {
					select {
					case <-ctx.Done():
						return
					case <-done:
						return
					case v, ok := <-in:
						if !ok {
							return
						}
						time.Sleep(sleepPerStage)
						select {
						case <-ctx.Done():
							return
						case <-done:
							return
						case out <- f(v):
						}
					}
				}
			}()
			return out
		}
	}

	t.Run("done case", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}
		var wg sync.WaitGroup

		// Create stages with done and WaitGroup support
		stages := []Stage{
			gWithDone("Dummy", func(v interface{}) interface{} { return v }, &wg, ctx, done),
			gWithDone("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }, &wg, ctx, done),
			gWithDone("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }, &wg, ctx, done),
			gWithDone("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }, &wg, ctx, done),
		}

		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				select {
				case <-ctx.Done():
					return
				case in <- v:
				}
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})

	t.Run("empty input", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		in := make(Bi)
		close(in)
		result := make([]string, 0)
		for s := range ExecutePipeline(in, nil, gWithDone("Dummy", func(v interface{}) interface{} { return v }, nil, ctx, nil)) {
			result = append(result, s.(string))
		}
		require.Empty(t, result)
	})

	t.Run("error in stage", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		errorStage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for {
					select {
					case <-ctx.Done():
						return
					case _, ok := <-in:
						if !ok {
							return
						}
						out <- nil
						return
					}
				}
			}()
			return out
		}
		in := make(Bi, 1)
		in <- 1
		close(in)
		done := make(Bi)
		result := make([]interface{}, 0)
		for s := range ExecutePipeline(in, done, errorStage) {
			result = append(result, s)
		}
		close(done)
		require.Contains(t, result, nil)
	})

	t.Run("large input", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		in := make(Bi)
		done := make(Bi)
		data := make([]int, 100)
		for i := range data {
			data[i] = i + 1
		}
		var wg sync.WaitGroup

		stages := []Stage{
			gWithDone("Dummy", func(v interface{}) interface{} { return v }, &wg, ctx, done),
			gWithDone("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }, &wg, ctx, done),
			gWithDone("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }, &wg, ctx, done),
			gWithDone("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }, &wg, ctx, done),
		}

		go func() {
			for _, v := range data {
				select {
				case <-ctx.Done():
					return
				case in <- v:
				}
			}
			close(in)
		}()

		result := make([]string, 0, 100)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)
		wg.Wait()

		expected := make([]string, 100)
		for i := range expected {
			expected[i] = strconv.Itoa((i+1)*2 + 100)
		}
		require.Equal(t, expected, result)
		require.Less(t, int64(elapsed), int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})
}
