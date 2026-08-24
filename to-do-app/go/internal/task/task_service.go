package task

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

type TaskService struct {
	taskList []*Task
}

func NewTaskService() *TaskService {
	return &TaskService{
		taskList: make([]*Task, 0),
	}
}

func (ts *TaskService) Add(id uint16, taskName string, createdAt time.Time, done bool) {
	ts.taskList = append(ts.taskList, newTask(id, taskName, createdAt, done))
}

func (ts *TaskService) List() []*Task {
	return ts.taskList
}

func (ts *TaskService) Complete(taskID uint16) error {
	for _, task := range ts.taskList {
		if task.id == taskID {
			task.complete()
			return nil
		}
	}
	return errors.New("taskID does not exist")
}

func (ts *TaskService) StdOutPrint() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	defer w.Flush()

	// Print Header
	fmt.Fprintln(w, "ID\tTASK NAME\tCREATED AT\tDONE")
	fmt.Fprintln(w, "--\t---------\t----------\t----")

	// Print Data Row
	for _, task := range ts.taskList {
		fmt.Fprintf(w, "%d\t%s\t%s\t%t\n", task.id, task.name, task.createdAt, task.done)
	}
}
