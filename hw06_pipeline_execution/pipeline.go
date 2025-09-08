package hw06_pipeline_execution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) Out

func pipeWithDone(src In, done In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-src:
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

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	current := in
	if len(stages) == 0 {
		return pipeWithDone(current, done)
	}

	for _, st := range stages {
		wrappedIn := pipeWithDone(current, done)
		current = st(wrappedIn)
	}

	return current
}
