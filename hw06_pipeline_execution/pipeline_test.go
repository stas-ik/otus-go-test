package hw06_pipeline_execution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
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
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
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
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})

	t.Run("empty input", func(t *testing.T) {
		in := make(Bi)
		close(in)
		result := make([]string, 0)
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		require.Empty(t, result)
	})

	t.Run("error in stage", func(t *testing.T) {
		errorStage := func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for range in {
					out <- nil
					return
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
		in := make(Bi)
		data := make([]int, 100)
		for i := range data {
			data[i] = i + 1
		}
		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 100)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		expected := make([]string, 100)
		for i := range expected {
			expected[i] = strconv.Itoa((i+1)*2 + 100)
		}
		require.Equal(t, expected, result)
		require.Less(t, int64(elapsed), int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})
}
