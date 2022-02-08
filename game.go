package wordle

import (
	"fmt"
	"math"
	"runtime"
)

type Game struct {
	dictionary []string
	p          player
}

type player interface {
	getGuess(bestGuess string) string
	getHint(guess string) wordHint
}

// NewGame creates a new game where the answer is unknown. A human is needed to type guesses into the Wordle game, and feed hints back into this program. Useful for playing real Wordles.
//
// In this mode, the solver offers what it thinks the best guess is, and you can choose to either follow that advice or use a different word.
func NewGame() *Game {
	return &Game{
		dictionary: ValidWords,
		p:          &humanPlayer{},
	}
}

// NewGameWithAnswer creates a new game where the answer is already known. Hints are self-calculated because the answer is known. Useful for seeing how the solver reacts to certain answers.
//
// In this mode, the solver always chooses the best guess.
func NewGameWithAnswer(answer string) *Game {
	return &Game{
		dictionary: ValidWords,
		p: computerPlayer{
			answer: answer,
		},
	}
}

// Play plays a game of Wordle. It returns the answer and the number of guesses needed to arrive at it.
//
// A game is played by repeatedly guessing. Each guess yields a hint, which narrows down the solution to a smaller set of potential words.
//
// For example, if a hint tells that the letter "u" is not present in a word, all words that have a "u" in them cannot be a solution.
//
// This process repeats until there is one word left - it is the answer.
//
// At each step, the best guess is chosen given the information revealed so far. See Game.getBestGuess for details.
func (g *Game) Play() (string, int) {
	guessCount := 1

	for len(g.dictionary) != 1 {

		if Verbose {
			fmt.Printf("(Guess #%v) Calculating best guess...\n", guessCount)
		}
		bestGuess, bestEntropy := g.getBestGuess(guessCount == 1)

		if Verbose {
			fmt.Printf("(Guess #%v) Best guess: %v (expected entropy: %v)\n", guessCount, bestGuess, bestEntropy)
		}

		guess := g.p.getGuess(bestGuess)
		hint := g.p.getHint(guess)

		if Verbose {
			fmt.Printf("(Guess #%v) Guess:      %v\n", guessCount, guess)
			fmt.Printf("(Guess #%v) Hint:       %v\n", guessCount, hint)
		}

		previousSize := len(g.dictionary)

		c := constraint{
			hint: hint,
			word: guess,
		}
		g.dictionary = c.filter(g.dictionary)

		if Verbose {
			fmt.Printf("(Guess #%v) Dict size:  %v -> %v (actual entropy: %v)\n", guessCount, previousSize, len(g.dictionary), math.Log2(float64(previousSize)/float64(len(g.dictionary))))
			fmt.Println()
		}

		if len(g.dictionary) == 0 {
			panic("That guess resulted in the dictionary being empty - no answer could be found. " +
				"If the answer is unknown, make sure the guess/hint were typed correctly. " +
				"If they were, or the answer is known, there's a bug somewhere.")
		}

		guessCount++
	}

	fmt.Println("Answer: ", g.dictionary[0])
	fmt.Println("Guesses:", guessCount)

	return g.dictionary[0], guessCount
}

// The worker pool used to calculate the entropy of potential guesses.
var workerPool = newEntropyWorkerPool(runtime.NumCPU())

// getBestGuess returns the best guess to make at this stage of the game.
//
// It does so by choosing the word which will narrow down the number of potential answers the most. In other words, the
// words which provides the most information. In other words: the words with the highest entropy.
//
// See entropyWorker.calculateEntropy for details on the entropy calculation.
//
// The first guess has no prior information, and thus is solely based on the dictionary of words.
// It also takes the longest to compute. So, it's calculated once and cached.
func (g *Game) getBestGuess(firstGuess bool) (string, float64) {
	if firstGuess {
		return "tares", 6.194052544375467
	}

	best, bestEntropy := "", 0.0

	for guessIndex, potentialGuess := range g.dictionary {
		info := workerPool.calculateEntropy(potentialGuess, g.dictionary)
		if Verbose {
			fmt.Printf("(%v/%v) %v: %v\n", guessIndex+1, len(g.dictionary), potentialGuess, info)
		}

		if info > bestEntropy {
			best = potentialGuess
			bestEntropy = info
		}
	}

	return best, bestEntropy
}
