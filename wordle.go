// Package wordle provides a Wordle solver. It can be used both when the answer is unknown (see NewGame) and
// when the answer is known (see NewGameWithAnswer).
package wordle

const (
	wordSize = 5
)

// Verbose controls the level of information printed to the console while playing a Game.
var Verbose = true
