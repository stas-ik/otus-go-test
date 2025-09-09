package hw06pipelineexecution

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
				go func() {
					for v := range src {
						_ = v
					}
				}()
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				select {
				case <-done:
					go func() {
						for vv := range src {
							_ = vv
						}
					}()
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
		stageOut := st(wrappedIn)
		current = pipeWithDone(stageOut, done)
	}

	return current
}
