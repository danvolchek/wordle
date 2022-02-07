package wordle

import (
	"fmt"
)

// A wordHint is a hint for an entire word.
type wordHint [wordSize]letterHint

// fromString parses this word hint from s, returning an error if s is invalid.
func (w *wordHint) fromString(s string) error {
	if len(s) != wordSize {
		return fmt.Errorf("wrong size: expected %v, got %v", wordSize, len(s))
	}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 'b':
			w[i] = absent
		case 'g':
			w[i] = correct
		case 'y':
			w[i] = present
		default:
			return fmt.Errorf("unexpected hint %v, use absent = b (black), present = y (yellow), correct = g (green)", string(s[i]))
		}
	}

	return nil
}

// A letterHint is a hint for a single letter. A letter is either absent from the word, present in the word but somewhere else,
// or correct and in the right position.
type letterHint int

const (
	absent letterHint = iota
	present
	correct
)

func (h letterHint) String() string {
	switch h {
	case absent:
		return "b"
	case present:
		return "y"
	case correct:
		return "g"
	default:
		panic(h)
	}
}

// createHint returns the hint associated with guess if the actual word is answer.
func createHint(guess, answer string) wordHint {
	// unscramble maps answer letter positions to the guess letter positions they correspond to
	unscramble := map[int]int{}
	for letterIndex := 0; letterIndex < wordSize; letterIndex++ {
		unscramble[letterIndex] = -1
	}

	for letterIndex := 0; letterIndex < wordSize; letterIndex++ {
		answerLetter := answer[letterIndex]
		guessLetter := guess[letterIndex]

		// if the guess letter matches the answer letter, the position is unchanged
		if guessLetter == answerLetter {
			unscramble[letterIndex] = letterIndex
		}
	}

	for letterIndex := 0; letterIndex < wordSize; letterIndex++ {
		answerLetter := answer[letterIndex]
		guessLetter := guess[letterIndex]

		// if the guess letter matches the answer letter, the position is another unused letter to move to
		if guessLetter != answerLetter {
			for letterIndex2 := 0; letterIndex2 < wordSize; letterIndex2++ {
				answerLetter2 := answer[letterIndex2]

				if answerLetter2 == guessLetter && unscramble[letterIndex2] == -1 {
					unscramble[letterIndex2] = letterIndex
					break
				}
			}
		}
	}

	// From the assignment of answer letters to guess letters, the hint can be created
	var hint wordHint
	for index, mapping := range unscramble {
		switch {
		case mapping == index: // the answer letter maps to the same position as the guess letter: the guess is correct
			hint[mapping] = correct
		case mapping != -1: // the answer letter maps to a different position in the guess: the guess is present
			hint[mapping] = present

			// in the default case, the answer letter has no mapping to the guess. The default value for wordHint is absent,
			// so doing nothing will keep that position absent
		}
	}

	return hint
}
