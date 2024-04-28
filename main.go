package main

import (
	"fmt"
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

func main() {
	seen := make(map[int64][]EkinState)
	currentStates := make([]EkinState, 1)
	currentStates[0] = EkinState{}

	addState := func(reached EkinState) {
		hash := reached.Hash()
		seen[hash] = append(seen[hash], reached)
	}

	isSeen := func(state EkinState) bool {
		hash := state.Hash()
		seenVecs, ok := seen[hash]
		if !ok {
			return false
		}

		for _, seenVec := range seenVecs {
			if seenVec == state {
				return true
			}
		}

		return false
	}

	addState(currentStates[0])

	lastMax := 0

	for i := 0; i < maxIterations; i++ {
		//t := time.Now()

		threadChan := make(chan int, maxThreads)
		outChan := make(chan []EkinState, len(currentStates))

		for _, currentReached := range currentStates {
			threadChan <- 1
			go func(reached EkinState) {
				nextStates := make([]EkinState, 0)
				for bitMask := 1; bitMask < (1 << n); bitMask++ {
					sum := 0
					for j := 0; j < n; j++ {
						if bitMask&(1<<j) != 0 {
							sum += reached[j]
						}
					}

					if sum < 0 {
						// next ekin state will add 1 from elements of the set
						nextState := reached
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState[j]++
							}
						}
						slices.Sort(nextState[:])

						if !isSeen(nextState) {
							nextStates = append(nextStates, nextState)
						}
					} else if sum > 0 {
						// next ekin state will subtract 1 from elements of the set
						nextState := reached
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState[j]--
							}
						}
						slices.Sort(nextState[:])

						if !isSeen(nextState) {
							nextStates = append(nextStates, nextState)
						}
					} else {
						// both operations are valid
						nextState1 := reached
						nextState2 := reached
						for j := 0; j < n; j++ {
							if bitMask&(1<<j) != 0 {
								nextState1[j]++
								nextState2[j]--
							}
						}
						slices.Sort(nextState1[:])
						slices.Sort(nextState2[:])

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

		nextStates := make([]EkinState, 0)
		for j := 0; j < len(currentStates); j++ {
			nextStates = append(nextStates, <-outChan...)
		}

		// filter out duplicates
		uniqueStates := make([]EkinState, 0)

		for _, state := range nextStates {
			if !isSeen(state) {
				uniqueStates = append(uniqueStates, state)
				addState(state)

				invState := state
				for j := 0; j < n; j++ {
					invState[j] *= -1
				}
				slices.Sort(invState[:])
				addState(invState)
			}
		}

		//fmt.Println("Iteration", i, "took", time.Since(t)/time.Duration(len(currentStates)), "per state", len(currentStates), "states", ", seen size:", len(seen))

		if len(uniqueStates) == 0 {
			break
		}

		//fmt.Println("Iteration", i, "Number of states", len(nextStates))
		for _, state := range uniqueStates {
			maxVal := 0
			for j := 0; j < n; j++ {
				if state[j] > maxVal {
					maxVal = state[j]
				}
			}
			if maxVal > lastMax {
				lastMax = maxVal
				fmt.Println("New max", lastMax, "at iteration", i)
			}
		}

		currentStates = uniqueStates

		if i >= maxIterations-1 {
			fmt.Println("Max iterations reached")
		}
	}

	fmt.Println("Done!")
}
