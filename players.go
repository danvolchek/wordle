package wordle

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// A humanPlayer plays a Game by:
// - manually typing the best guess into the game (shown through stdout)
// - entering the resulting hint through stdin
type humanPlayer struct{}

func (h humanPlayer) getGuess(bestGuess string) string {
	if !Verbose {
		fmt.Println("Best guess:", bestGuess)
	}

	for {
		result := readLine("Guess")
		if len(result) == 0 {
			return bestGuess
		}

		if len(result) != wordSize {
			fmt.Printf("Bad guess: wrong size: expected %v, got %v\n", wordSize, len(result))
			continue
		}

		var hint wordHint
		if hint.fromString(result) == nil {
			fmt.Println("Bad guess: that looks like a hint, not a guess")
			continue
		}

		return result
	}
}

func (h humanPlayer) getHint(guess string) wordHint {
	var hint wordHint
	for {
		result := readLine("Hint")

		err := hint.fromString(result)
		if err == nil {
			return hint
		}

		fmt.Printf("Bad hint: %v\n", err)
	}
}

func readLine(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + ": ")
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(text)
}

// A computerPlayer plays a Game by:
// - using the best guess
// - calculating the hint by comparing against the answer
type computerPlayer struct {
	answer string
}

func (c computerPlayer) getGuess(bestGuess string) string {
	return bestGuess
}

func (c computerPlayer) getHint(guess string) wordHint {
	return createHint(guess, c.answer)
}
