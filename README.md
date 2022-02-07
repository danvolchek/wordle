# Wordle

## Overview

This repo contains a parallel Wordle solver. It works by choosing guesses that reduce the number of possible solution words the most.
It does so by guessing words which reveal the most information about the solution. In other words, it guesses the words with highest entropy.
Entropy of words is calculated using as many logical cores as are available on the machine.

See [game.go](game.go) (specifically `Play` and `getBestGuess`) for how a game is played (i.e. how the entropy of words is used)
and [entropy.go](entropy.go) (specifically `calculateEntropy`) for how word entropy is calculated.

## Usage

See [main.go](main/main.go). In a nutshell,

```go
package main

import "github.com/danvolchek/wordle"

func main() {
	// solve a Wordle where the answer is unknown (e.g. current day)
	wordle.NewGame().Play()
}
```