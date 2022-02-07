package wordle

// A constraint is the combination of a word hint and a word. Words can be tested to see if they satisfy the constraint -
// i.e. whether a word is possible given the known hint.
type constraint struct {
	hint wordHint
	word string
}

// satisfies returns whether word meets all the constraints described by c.
func (c constraint) satisfies(word string) bool {
	// Using the constraint's word as the guess, and word as the answer, if the resulting hint is the same as the
	// constraint's hint, then word satisfies the constraint. In other words, it means that word is possibly the answer.
	return createHint(c.word, word) == c.hint
}

// filter returns the subset of words in dictionary which satisfy c.
func (c constraint) filter(dictionary []string) []string {
	var result []string

	for _, word := range dictionary {
		if c.satisfies(word) {
			result = append(result, word)
		}
	}

	return result
}

// filterNum returns the size of the subset of words in dictionary which satisfy c.
func (c constraint) filterNum(dictionary []string) int {
	result := 0

	for _, word := range dictionary {
		if c.satisfies(word) {
			result += 1
		}
	}

	return result
}
