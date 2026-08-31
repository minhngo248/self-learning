package task

import "time"

type Task struct {
	id        uint16
	name      string
	createdAt time.Time
	done      bool
}

func newTask(id uint16, name string, createdAt time.Time, done bool) *Task {
	return &Task{
		id:        id,
		name:      name,
		createdAt: createdAt,
		done:      done,
	}
}

func (t *Task) complete() {
	t.done = true
}

func (t *Task) GetID() uint16 {
	return t.id
}

func (t *Task) GetName() string {
	return t.name
}

func (t *Task) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t *Task) IsDone() bool {
	return t.done
}
