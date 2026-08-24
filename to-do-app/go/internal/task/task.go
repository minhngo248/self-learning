package task

import "time"

type Task struct {
	id        uint16
	name      string
	createdAt time.Time
	done      bool
}

func NewTask(id uint16, name string) *Task {
	return &Task{
		id:        id,
		name:      name,
		createdAt: time.Now(),
		done:      false,
	}
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

func (t *Task) GetDone() bool {
	return t.done
}

func (t *Task) Complete() {
	t.done = true
}
