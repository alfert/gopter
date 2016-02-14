package gopter

import "math"

// Prop represent some kind of property that (drums please) can and should be checked
type Prop func(*GenParameters) *PropResult

// Check the property using specific parameters
func (prop Prop) Check(parameters *TestParameters) *TestResult {
	iterations := math.Ceil(float64(parameters.MinSuccessfulTests) / float64(parameters.Workers))
	sizeStep := float64(parameters.MaxSize-parameters.MinSize) / (iterations * float64(parameters.Workers))

	genParameters := GenParameters{
		Rng: parameters.Rng,
	}
	runner := &runner{
		parameters: parameters,
		worker: func(workerIdx int, shouldStop shouldStop) *TestResult {
			var n int
			var d int

			isExhaused := func() bool {
				return n+d > parameters.MinSuccessfulTests &&
					float64(1+parameters.Workers*n)*parameters.MaxDiscardRatio < float64(d)
			}

			for !shouldStop() && n < int(iterations) {
				size := float64(parameters.MinSize) + (sizeStep * float64(workerIdx+(parameters.Workers*(n+d))))
				propResult := prop(genParameters.WithSize(int(size)))

				switch propResult.Status {
				case PropUndecided:
					d++
					if isExhaused() {
						return &TestResult{
							Status:    TestExhausted,
							Succeeded: n,
							Discarded: d,
						}
					}
				case PropTrue:
					n++
				case PropProof:
					n++
					return &TestResult{
						Status:    TestProved,
						Succeeded: n,
						Discarded: d,
					}
				case PropFalse:
					return &TestResult{
						Status:    TestFailed,
						Succeeded: n,
						Discarded: d,
					}
				case PropError:
					return &TestResult{
						Status:    TestError,
						Succeeded: n,
						Discarded: d,
					}
				}
			}

			if isExhaused() {
				return &TestResult{
					Status:    TestExhausted,
					Succeeded: n,
					Discarded: d,
				}
			}
			return &TestResult{
				Status:    TestPassed,
				Succeeded: n,
				Discarded: d,
			}
		},
	}

	return runner.runWorkers()
}
