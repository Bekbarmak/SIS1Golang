package task

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Processor interface {
	Process(input string) (Result, error)
}

type TextAnalyzer struct{}

func (TextAnalyzer) Process(input string) (Result, error) {
	words := strings.FieldsFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	unique := make(map[string]struct{}, len(words))
	longest := ""
	longestLen := 0
	for _, w := range words {
		unique[strings.ToLower(w)] = struct{}{}
		if n := utf8.RuneCountInString(w); n > longestLen {
			longest = w
			longestLen = n
		}
	}

	return Result{
		Words:       len(words),
		Characters:  utf8.RuneCountInString(input),
		Lines:       strings.Count(strings.TrimSpace(input), "\n") + 1,
		UniqueWords: len(unique),
		LongestWord: longest,
	}, nil
}
