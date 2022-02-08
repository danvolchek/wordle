package wordle

import (
	"math"
)

// An entropyWorker calculates the entropy of words for a given list of hints.
type entropyWorker struct {
	jobs      <-chan entropyWorkJob
	result    chan<- entropyWorkResult
	workerNum int
	hints     []wordHint
}

type entropyWorkJob struct {
	word       string
	dictionary []string
}

// An entropyWorkResult is the result of an entropy calculation by an entropyWorker.
type entropyWorkResult struct {
	workerNum int
	entropy   float64
}

func (e entropyWorker) work() {
	for {
		select {
		case job := <-e.jobs:
			e.calculateEntropy(job.word, job.dictionary)
		}
	}
}

// calculateEntropy calculates the entropy for the given word in the context of a dictionary of possible words using the hints configured for this worker.
//
// Note: this is based on https://www.youtube.com/watch?v=v68zYyaEmEA and https://en.wikipedia.org/wiki/Entropy_(information_theory).
//
// The entropy of a word is a measure of the information it provides. Given a word x and dictionary d, an entropy of e means that on average guessing x will reduce the size of d by 2**e
//
// For example, if "foo" had an entropy of 1, on average guessing it would reduce the number of possible words by 2 (e.g. 100 -> 50).
// If it had an entropy of 2, guessing it would reduce the number of possible words by 4 (e.g. 100 -> 25). 3 would reduce by 8 (100 -> 12.5), and so on.
//
// The entropy of a word is calculated as the sum of the expected information for all possible hints it results in.
// (An aside: the answer is unknown, so all hints are possible. Some yield 0 remaining words in the dictionary; they are incorrect hints and ignored.
// Entropy is additive - this worker is configured to calculate a subset of the hints which it adds together before sending to the pool, which does a final addition to yield the final entropy).
//
// The expected information of a hint is how likely that hint is (if it's less likely it contributes less to the average)
// multiplied by how much information it provides (if it's less likely, it provides more information, because on being correct it reduces the number of possibilities more).
//
// How likely a hint is defined as the number of remaining valid words after applying the hint to the dictionary, divided by the total words. If more words are left, it's more likely the answer is one of those words.
// How much information a hint provides is defined as log2(hint likeliness), because of fancy information theory.
//
// For example, if the dictionary contains 2 words "bar" and "baz", the possible hints are "ggg" and "bgg" for both words (the cases where either is the answer).
// The expected information of "ggg" is (1/2) * log2(1/(1/2)) = 0.5 * log2(2) = 0.5. It's the same for "bgg", yielding an entropy of 0.5 + 0.5 = 1 for both words.
// This means guessing either will reduce the dictionary size from 2 to 2/(2**1) = 1, yielding the answer, as expected.
//
// It's not always the case that each hint yields the same information (the above is a simple case), and so the information gained from a guess can be more or less than the expected information, depending on which hint
// actually occurred.
//
// Multiplying these two together, and summing across all hints, yields the entropy for a word.
func (e entropyWorker) calculateEntropy(word string, dictionary []string) {
	dictionarySize := float64(len(dictionary))

	var entropy float64

	for _, hint := range e.hints {
		c := constraint{
			hint: hint,
			word: word,
		}

		remainingSize := float64(c.filterNum(dictionary))

		if remainingSize == 0 {
			continue
		}

		probability := remainingSize / dictionarySize
		entropy += math.Log2(1/probability) * probability
	}

	e.result <- entropyWorkResult{
		workerNum: e.workerNum,
		entropy:   entropy,
	}
}

// An entropyWorkerPool calculates the entropy of a given word using a pool of workers to maximize resource utilization.
// Entropy is the measure used to determine quality of words.
// The pool shards the possible hints across all of its workers, parallelizing the work.
type entropyWorkerPool struct {
	numWorkers int

	workers []chan entropyWorkJob
	results chan entropyWorkResult
	done    chan bool
}

// newEntropyWorkerPool creates an entropyWorkerPool with the configured number of workers.
func newEntropyWorkerPool(numWorkers int) entropyWorkerPool {
	wp := entropyWorkerPool{
		numWorkers: numWorkers,
		workers:    make([]chan entropyWorkJob, numWorkers),
		results:    make(chan entropyWorkResult, numWorkers),
		done:       make(chan bool),
	}

	hintsPerWorker := (numPossibleWordHints / numWorkers) + 1

	for workerNum := 0; workerNum < numWorkers; workerNum++ {
		startHint := workerNum * hintsPerWorker
		stopHint := (workerNum + 1) * hintsPerWorker
		if stopHint > numPossibleWordHints {
			stopHint = numPossibleWordHints
		}

		jobChan := make(chan entropyWorkJob)
		wp.workers[workerNum] = jobChan

		worker := entropyWorker{
			jobs:      jobChan,
			result:    wp.results,
			workerNum: workerNum,
			hints:     possibleWordHints[startHint:stopHint],
		}

		go worker.work()
	}

	return wp
}

// collectWorkerResults waits for all workers to complete and then aggregates their results into a final entropy
// result. It does so in a deterministic manner so that race conditions between worker completion and floating point math
// don't cause non-deterministic results.
func (e entropyWorkerPool) collectWorkerResults() float64 {
	results := make([]float64, e.numWorkers)

	go func() {
		for workerNum := 0; workerNum < e.numWorkers; workerNum++ {
			result := <-e.results
			results[result.workerNum] = result.entropy
		}
		e.done <- true
	}()

	<-e.done

	// The entropy of the word is the sum of the entropy of all the workers.
	sum := 0.0
	for workerNum := 0; workerNum < len(results); workerNum++ {
		sum += results[workerNum]
	}

	return sum
}

// calculateEntropy starts the pool's workers on the task of calculating the entropy for the given word in context of
// the given dictionary.
func (e entropyWorkerPool) calculateEntropy(word string, dictionary []string) float64 {
	// start workers
	for _, worker := range e.workers {
		worker <- entropyWorkJob{
			word:       word,
			dictionary: dictionary,
		}
	}

	return e.collectWorkerResults()
}

var (
	possibleWordHints    = allPossibleWordHints()
	numPossibleWordHints = len(possibleWordHints)
)

// allPossibleWordHints returns all 3**5 possible permutations of the three possible letter combined for 5 words
func allPossibleWordHints() []wordHint {
	result := make([]wordHint, int(math.Pow(3, wordSize)))

	var current wordHint

	for i := 0; i < len(result); i++ {
		result[i] = current

		index := 0
		for current[index] == correct {
			current[index] = absent
			index++

			if index == wordSize {
				return result
			}
		}
		current[index] += 1
	}

	panic("didn't fill result - is 3 the right number of letter hints?")
}
