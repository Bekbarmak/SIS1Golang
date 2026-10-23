package task

import (
	"cmp"
	"fmt"
	"slices"
)

type Repository interface {
	Add(t Task) Task
	Get(id int) (Task, error)
	Update(t Task) error
	List() []Task
}

type MemoryRepository struct {
	tasks  map[int]Task
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (r *MemoryRepository) Add(t Task) Task {
	t.ID = r.nextID
	r.nextID++
	r.tasks[t.ID] = t.clone()
	return t.clone()
}

func (r *MemoryRepository) Get(id int) (Task, error) {
	t, ok := r.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("%w: id %d", ErrTaskNotFound, id)
	}
	return t.clone(), nil
}

func (r *MemoryRepository) Update(t Task) error {
	if _, ok := r.tasks[t.ID]; !ok {
		return fmt.Errorf("%w: id %d", ErrTaskNotFound, t.ID)
	}
	r.tasks[t.ID] = t.clone()
	return nil
}

func (r *MemoryRepository) List() []Task {
	out := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t.clone())
	}
	slices.SortFunc(out, func(a, b Task) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return out
}
