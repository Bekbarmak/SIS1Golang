package task

import "fmt"

type Service struct {
	repo      Repository
	processor Processor
}

func NewService(repo Repository, processor Processor) *Service {
	return &Service{repo: repo, processor: processor}
}

func (s *Service) Create(input string) (Task, error) {
	if err := Validate(input); err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return s.repo.Add(Task{Input: input, State: StatePending}), nil
}

func (s *Service) Get(id int) (Task, error) {
	return s.repo.Get(id)
}

func (s *Service) List() []Task {
	return s.repo.List()
}

func (s *Service) Process(id int) (Task, error) {
	t, err := s.repo.Get(id)
	if err != nil {
		return Task{}, err
	}

	if err := t.transition(StateProcessing); err != nil {
		return t, fmt.Errorf("process task %d: %w", id, err)
	}
	if err := s.repo.Update(t); err != nil {
		return t, err
	}

	result, procErr := s.processor.Process(t.Input)
	if procErr != nil {
		if err := t.transition(StateFailed); err != nil {
			return t, err
		}
		if err := s.repo.Update(t); err != nil {
			return t, err
		}
		return t, fmt.Errorf("process task %d: %w", id, procErr)
	}

	t.Result = &result
	if err := t.transition(StateCompleted); err != nil {
		return t, err
	}
	if err := s.repo.Update(t); err != nil {
		return t, err
	}
	return t.clone(), nil
}
