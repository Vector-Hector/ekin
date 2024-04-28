package main

import (
	"fmt"
	"github.com/google/uuid"
	"slices"
)

const n = 6
const maxIterations = 10000000
const maxThreads = 8

type EkinState [n]int

func (e EkinState) Hash() int64 {
	hash := int64(0)
	for i := 0; i < n; i++ {
		hash = hash*31 + int64(e[i])
	}
	return hash
}

func (e EkinState) ToString() string {
	return fmt.Sprint(e)
}

type ReachedEkinState struct {
	vec EkinState
	id  string
}

func main() {
	seen := make(map[int64][]EkinState)
	currentStates := make([]*ReachedEkinState, 1)
	currentStates[0] = &ReachedEkinState{
		vec: EkinState{}, // zero filled ekin state
		id:  uuid.NewString(),
	}

	addState := func(reached *ReachedEkinState) {
		hash := reached.vec.Hash()
		seen[hash] = append(seen[hash], reached.vec)
	}

	isSeen := func(state *ReachedEkinState) bool {
		hash := state.vec.Hash()
		seenVecs, ok := seen[hash]
		if !ok {
			return false
		}

		for _, seenVec := range seenVecs {
			if seenVec == state.vec {
				return true
			}
		}

		return false
	}

	addState(currentStates[0])

	prevStateLogger, err := NewCsvLogger(fmt.Sprintf("prev_state_%d.csv", n))
	if err != nil {
		panic(err)
	}
	defer prevStateLogger.Close()

	prevStateLogger.MustLog([]string{"id", "vec", "previous"})
	prevStateLogger.MustLog([]string{currentStates[0].id, currentStates[0].vec.ToString(), ""})

	lastMax := 0

	for i := 0; i < maxIterations; i++ {
		//t := time.Now()

		threadChan := make(chan int, maxThreads)
		outChan := make(chan []*ReachedEkinState, len(currentStates))

		for _, currentReached := range currentStates {
			threadChan <- 1
			go func(reached *ReachedEkinState) {
				nextStates := make([]*ReachedEkinState, 0)
				for bitMask := 1; bitMask < (1 << n); bitMask++ {
					sum := 0
					for j := 0; j < n; j++ {
						if bitMask&(1<<j) != 0 {
							sum += reached.vec[j]
						}
					}

					if sum < 0 {
						// next ekin state will add 1 from elements of the set
						nextState := &ReachedEkinState{
							vec: reached.vec,
							id:  uuid.NewString(),
						}
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState.vec[j]++
							}
						}
						slices.Sort(nextState.vec[:])

						if !isSeen(nextState) {
							nextStates = append(nextStates, nextState)
						}
					} else if sum > 0 {
						// next ekin state will subtract 1 from elements of the set
						nextState := &ReachedEkinState{
							vec: reached.vec,
							id:  uuid.NewString(),
						}
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState.vec[j]--
							}
						}
						slices.Sort(nextState.vec[:])

						if !isSeen(nextState) {
							nextStates = append(nextStates, nextState)
						}
					} else {
						// both operations are valid
						nextState1 := &ReachedEkinState{
							vec: reached.vec,
							id:  uuid.NewString(),
						}
						nextState2 := &ReachedEkinState{
							vec: reached.vec,
							id:  uuid.NewString(),
						}
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState1.vec[j]++
								nextState2.vec[j]--
							}
						}
						slices.Sort(nextState1.vec[:])
						slices.Sort(nextState2.vec[:])

						if !isSeen(nextState1) {
							nextStates = append(nextStates, nextState1)
						}
					}
				}
				outChan <- nextStates
				<-threadChan
			}(currentReached)
		}

		for j := 0; j < maxThreads; j++ {
			threadChan <- 1
		}

		nextStates := make([]*ReachedEkinState, 0)
		for j := 0; j < len(currentStates); j++ {
			nextStates = append(nextStates, <-outChan...)
		}

		// filter out duplicates
		uniqueStates := make([]*ReachedEkinState, 0)

		for _, state := range nextStates {
			if !isSeen(state) {
				uniqueStates = append(uniqueStates, state)
				addState(state)
			}
		}

		nextStates = uniqueStates

		//fmt.Println("Iteration", i, "took", time.Since(t)/time.Duration(len(currentStates)), "per state", len(currentStates), "states")

		if len(nextStates) == 0 {
			break
		}

		//fmt.Println("Iteration", i, "Number of states", len(nextStates))
		for _, state := range nextStates {
			maxVal := 0
			for j := 0; j < n; j++ {
				if state.vec[j] > maxVal {
					maxVal = state.vec[j]
				}
			}
			if maxVal > lastMax {
				lastMax = maxVal
				fmt.Println("New max", lastMax, "at iteration", i)
			}
		}

		currentStates = nextStates

		if i >= maxIterations-1 {
			fmt.Println("Max iterations reached")
		}
	}

	fmt.Println("Done!")
}
