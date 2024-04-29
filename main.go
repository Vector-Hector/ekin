package main

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

const n = 3
const maxIterations = 10000000
const maxThreads = 8

type BaseInt int8

const baseIntMax = math.MaxInt8
const baseIntMin = math.MinInt8

type HashInt int32

type EkinState [n]BaseInt

func (e EkinState) Hash() HashInt {
	hash := HashInt(0)
	for i := 0; i < n; i++ {
		hash = hash*31 + HashInt(e[i])
	}
	return hash
}

func (e EkinState) ToString() string {
	return fmt.Sprint(e)
}

func main() {
	seen := make(map[HashInt][]EkinState)
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

	logger, err := NewCsvLogger(fmt.Sprintf("data/size-%d-iterations.csv", n))
	if err != nil {
		panic(err)
	}

	defer logger.Close()

	logger.MustLog([]string{"new_max", "iterations"})

	lastMax := BaseInt(0)

	for i := 0; i < maxIterations; i++ {
		//t := time.Now()

		threadChan := make(chan int, maxThreads)
		outChan := make(chan []EkinState, len(currentStates))

		for _, currentReached := range currentStates {
			threadChan <- 1
			go func(reached EkinState) {
				nextStates := make([]EkinState, 0)
				for bitMask := 1; bitMask < (1 << n); bitMask++ {
					sum := int64(0)
					for j := 0; j < n; j++ {
						if bitMask&(1<<j) != 0 {
							sum += int64(reached[j])
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

						if !isSeen(nextState2) {
							nextStates = append(nextStates, nextState2)
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

		// filter out duplicates
		uniqueStates := make([]EkinState, 0)

		for j := 0; j < len(currentStates); j++ {
			nextStates := <-outChan

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
					uniqueStates = append(uniqueStates, invState)
				}
			}
		}

		//fmt.Println("Iteration", i, "took", time.Since(t)/time.Duration(len(currentStates)), "per state", len(currentStates), "states", ", seen size:", len(seen))

		if len(uniqueStates) == 0 {
			break
		}

		//fmt.Println("Iteration", i, "Number of states", len(nextStates))
		for _, state := range uniqueStates {
			maxVal := BaseInt(0)
			for j := 0; j < n; j++ {
				if state[j] > maxVal {
					maxVal = state[j]
				}

				if state[j] == baseIntMax || state[j] == baseIntMin {
					panic(errors.New(fmt.Sprintf("base int overflow (reached %d)", state[j])))
				}
			}
			if maxVal > lastMax {
				lastMax = maxVal
				fmt.Println("New max", lastMax, "at iteration", i)
				logger.MustLog([]string{fmt.Sprint(lastMax), fmt.Sprint(i)})
			}
		}

		currentStates = uniqueStates

		if i >= maxIterations-1 {
			fmt.Println("Max iterations reached")
		}
	}

	fmt.Println("Done!")
}
