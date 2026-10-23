package task

import (
	"errors"
	"testing"
)

type failingProcessor struct{}

func (failingProcessor) Process(string) (Result, error) {
	return Result{}, errors.New("processor failure")
}

func newTestService() *Service {
	return NewService(NewMemoryRepository(), TextAnalyzer{})
}

func TestCreateTask(t *testing.T) {
	svc := newTestService()

	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("ID = %d, want 1", created.ID)
	}
	if created.State != StatePending {
		t.Errorf("State = %s, want %s", created.State, StatePending)
	}
	if created.Input != "Hello Go world" {
		t.Errorf("Input = %q, want %q", created.Input, "Hello Go world")
	}
	if created.Result != nil {
		t.Errorf("Result = %+v, want nil before processing", created.Result)
	}
}

func TestCreateTaskInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"empty", "", ErrEmptyInput},
		{"spaces only", "     ", ErrEmptyInput},
		{"whitespace mix", " \n\t \r\n", ErrEmptyInput},
		{"invalid utf8", "\xff\xfe", ErrInvalidEncoding},
		{"punctuation only", "!!! ... ???", ErrNoWords},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			_, err := svc.Create(tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if got := len(svc.List()); got != 0 {
				t.Errorf("stored tasks = %d, want 0", got)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("find me")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != created.ID || got.Input != "find me" || got.State != StatePending {
		t.Errorf("Get = %+v, want task %+v", got, created)
	}
}

func TestGetNonexistentTask(t *testing.T) {
	svc := newTestService()
	_, err := svc.Get(42)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestProcessTask(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	processed, err := svc.Process(created.ID)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if processed.State != StateCompleted {
		t.Errorf("returned State = %s, want %s", processed.State, StateCompleted)
	}
	if processed.Result == nil {
		t.Fatal("Result is nil after processing")
	}

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if stored.State != StateCompleted {
		t.Errorf("stored State = %s, want %s", stored.State, StateCompleted)
	}
	if stored.Result == nil || *stored.Result != *processed.Result {
		t.Errorf("stored Result = %+v, want %+v", stored.Result, processed.Result)
	}
}

func TestProcessResultCorrectness(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Result
	}{
		{
			name:  "simple sentence",
			input: "Hello Go world",
			want:  Result{Words: 3, Characters: 14, Lines: 1, UniqueWords: 3, LongestWord: "Hello"},
		},
		{
			name:  "repeated words",
			input: "the cat and the hat",
			want:  Result{Words: 5, Characters: 19, Lines: 1, UniqueWords: 4, LongestWord: "the"},
		},
		{
			name:  "multiple lines and case",
			input: "Line one\nline TWO",
			want:  Result{Words: 4, Characters: 17, Lines: 2, UniqueWords: 3, LongestWord: "Line"},
		},
		{
			name:  "unicode text",
			input: "Привет, мир!",
			want:  Result{Words: 2, Characters: 12, Lines: 1, UniqueWords: 2, LongestWord: "Привет"},
		},
		{
			name:  "punctuation separates words",
			input: "well-known, state-of-the-art",
			want:  Result{Words: 6, Characters: 28, Lines: 1, UniqueWords: 6, LongestWord: "known"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			created, err := svc.Create(tt.input)
			if err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
			processed, err := svc.Process(created.ID)
			if err != nil {
				t.Fatalf("Process returned error: %v", err)
			}
			if processed.Result == nil {
				t.Fatal("Result is nil")
			}
			if *processed.Result != tt.want {
				t.Errorf("Result = %+v, want %+v", *processed.Result, tt.want)
			}
		})
	}
}

func TestProcessNonexistentTask(t *testing.T) {
	svc := newTestService()
	_, err := svc.Process(7)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrTaskNotFound)
	}
}

func TestProcessCompletedTaskAgain(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("process me once")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	first, err := svc.Process(created.ID)
	if err != nil {
		t.Fatalf("first Process returned error: %v", err)
	}

	_, err = svc.Process(created.ID)
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidTransition)
	}

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if stored.State != StateCompleted {
		t.Errorf("State = %s, want %s", stored.State, StateCompleted)
	}
	if stored.Result == nil || *stored.Result != *first.Result {
		t.Errorf("Result changed after rejected second processing: %+v", stored.Result)
	}
}

func TestProcessFailure(t *testing.T) {
	svc := NewService(NewMemoryRepository(), failingProcessor{})
	created, err := svc.Create("this will fail")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := svc.Process(created.ID); err == nil {
		t.Fatal("Process returned nil error, want failure")
	}

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if stored.State != StateFailed {
		t.Errorf("State = %s, want %s", stored.State, StateFailed)
	}
	if stored.Result != nil {
		t.Errorf("Result = %+v, want nil", stored.Result)
	}

	if _, err := svc.Process(created.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("reprocessing failed task: error = %v, want %v", err, ErrInvalidTransition)
	}
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		from State
		to   State
		want bool
	}{
		{StatePending, StateProcessing, true},
		{StateProcessing, StateCompleted, true},
		{StateProcessing, StateFailed, true},
		{StatePending, StateCompleted, false},
		{StatePending, StateFailed, false},
		{StateCompleted, StateProcessing, false},
		{StateCompleted, StatePending, false},
		{StateFailed, StateProcessing, false},
		{StateProcessing, StatePending, false},
	}

	for _, tt := range tests {
		name := string(tt.from) + "_to_" + string(tt.to)
		t.Run(name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("CanTransitionTo = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultipleIndependentTasks(t *testing.T) {
	svc := newTestService()
	inputs := []string{
		"alpha beta",
		"gamma",
		"one two three four",
	}

	ids := make([]int, 0, len(inputs))
	seen := make(map[int]bool)
	for _, in := range inputs {
		created, err := svc.Create(in)
		if err != nil {
			t.Fatalf("Create(%q) returned error: %v", in, err)
		}
		if seen[created.ID] {
			t.Fatalf("duplicate ID %d", created.ID)
		}
		seen[created.ID] = true
		ids = append(ids, created.ID)
	}

	first, err := svc.Process(ids[0])
	if err != nil {
		t.Fatalf("Process(%d) returned error: %v", ids[0], err)
	}
	third, err := svc.Process(ids[2])
	if err != nil {
		t.Fatalf("Process(%d) returned error: %v", ids[2], err)
	}

	if first.Result.Words != 2 {
		t.Errorf("first Words = %d, want 2", first.Result.Words)
	}
	if third.Result.Words != 4 {
		t.Errorf("third Words = %d, want 4", third.Result.Words)
	}

	second, err := svc.Get(ids[1])
	if err != nil {
		t.Fatalf("Get(%d) returned error: %v", ids[1], err)
	}
	if second.State != StatePending {
		t.Errorf("second State = %s, want %s", second.State, StatePending)
	}
	if second.Result != nil {
		t.Errorf("second Result = %+v, want nil", second.Result)
	}
	if second.Input != "gamma" {
		t.Errorf("second Input = %q, want %q", second.Input, "gamma")
	}

	all := svc.List()
	if len(all) != len(inputs) {
		t.Fatalf("List length = %d, want %d", len(all), len(inputs))
	}
	for i, tk := range all {
		if tk.Input != inputs[i] {
			t.Errorf("List[%d].Input = %q, want %q", i, tk.Input, inputs[i])
		}
	}
}

func TestReturnedTaskIsACopy(t *testing.T) {
	svc := newTestService()
	created, err := svc.Create("original text")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	created.Input = "tampered"
	created.State = StateCompleted

	stored, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if stored.Input != "original text" || stored.State != StatePending {
		t.Errorf("stored task was modified through returned value: %+v", stored)
	}
}
