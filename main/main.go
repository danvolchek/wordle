package main

import (
	"github.com/danvolchek/wordle"
	"math/rand"
	"time"
)

const (
	// Set to true to play a Wordle where you don't know the answer, e.g. the current day.
	// Set to false to play a Wordle where the answer is known (chosen randomly).
	unknownWord = true
)

func main() {
	if unknownWord {
		wordle.NewGame().Play()
	} else {
		rand.Seed(time.Now().Unix())
		randomWord := rand.Intn(2315)
		wordle.NewGameWithAnswer(wordle.ValidWords[randomWord]).Play()
	}
}
