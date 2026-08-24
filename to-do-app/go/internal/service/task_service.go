package service

import (
	"errors"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
)

type TaskService struct {
	current  uint16
	taskList map[uint16]task.Task
}

func NewTaskService() *TaskService {
	return &TaskService{
		taskList: make(map[uint16]task.Task),
	}
}

func (ts *TaskService) Add(taskName string) {
	ts.taskList[ts.current] = *task.NewTask(ts.current, taskName)
	ts.current++
}

func (ts *TaskService) List() []task.Task {
	tasks := make([]task.Task, 0)
	for _, task := range ts.taskList {
		tasks = append(tasks, task)
	}
	return tasks
}

func (ts *TaskService) Complete(taskID uint16) error {
	if _, ok := ts.taskList[taskID]; !ok {
		return errors.New("taskID does not exist")
	}
	task := ts.taskList[taskID]
	task.Complete()
	ts.taskList[taskID] = task
	return nil
}
