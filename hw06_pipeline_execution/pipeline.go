package hw06_pipeline_execution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) Out

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := make(Bi)
	current := in

	for _, stage := range stages {
		next := make(Bi)
		go func(inCh In, st Stage) {
			defer close(next)
			for {
				select {
				case <-done:
					return
				case v, ok := <-inCh:
					if !ok {
						return
					}
					tempCh := make(Bi)
					go func() {
						defer close(tempCh)
						tempCh <- v
					}()
					outCh := st(tempCh)
					for val := range outCh {
						select {
						case <-done:
							return
						case next <- val:
						}
					}
				}
			}
		}(current, stage)
		current = next
	}

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-current:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- v:
				}
			}
		}
	}()

	return out
}
