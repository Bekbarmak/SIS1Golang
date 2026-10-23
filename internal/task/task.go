package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrEmptyInput        = errors.New("input text is empty")
	ErrInvalidEncoding   = errors.New("input text is not valid UTF-8")
	ErrNoWords           = errors.New("input text contains no letters or digits")
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTransition = errors.New("invalid state transition")
)

type State string

const (
	StatePending    State = "pending"
	StateProcessing State = "processing"
	StateCompleted  State = "completed"
	StateFailed     State = "failed"
)

var allowedTransitions = map[State][]State{
	StatePending:    {StateProcessing},
	StateProcessing: {StateCompleted, StateFailed},
}

func (s State) CanTransitionTo(next State) bool {
	return slices.Contains(allowedTransitions[s], next)
}

type Result struct {
	Words       int
	Characters  int
	Lines       int
	UniqueWords int
	LongestWord string
}

type Task struct {
	ID     int
	Input  string
	State  State
	Result *Result
}

func (t *Task) transition(next State) error {
	if !t.State.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, t.State, next)
	}
	t.State = next
	return nil
}

func (t Task) clone() Task {
	if t.Result != nil {
		r := *t.Result
		t.Result = &r
	}
	return t
}

func Validate(input string) error {
	if strings.TrimSpace(input) == "" {
		return ErrEmptyInput
	}
	if !utf8.ValidString(input) {
		return ErrInvalidEncoding
	}
	hasContent := strings.ContainsFunc(input, func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	})
	if !hasContent {
		return ErrNoWords
	}
	return nil
}
