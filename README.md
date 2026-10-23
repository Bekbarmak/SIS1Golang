# SIS Project - Text Analyzer

A task is a piece of text. Processing a task analyzes the text and produces a result: word count, character count, line count, unique word count, and the longest word.

## Run

    go run ./cmd/sis
    go test ./...

## Task lifecycle

    pending -> processing -> completed
                         \-> failed

| From       | Allowed next state     |
|------------|------------------------|
| pending    | processing             |
| processing | completed, failed      |
| completed  | none                   |
| failed     | none                   |

Any other transition returns `ErrInvalidTransition`. Processing a completed or failed task again is rejected.

## Validation

Input is rejected if it is empty or whitespace-only, is not valid UTF-8, or contains no letters or digits.

## Layout

- `internal/task/task.go` - Task, State, Result, errors, validation
- `internal/task/processor.go` - Processor interface and TextAnalyzer
- `internal/task/repository.go` - Repository interface and in-memory implementation
- `internal/task/service.go` - Create, Get, Process, List
- `cmd/sis` - demo runner