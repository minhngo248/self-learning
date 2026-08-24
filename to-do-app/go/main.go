package main

import (
	"fmt"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/service"
	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
)

func main() {
	var choice int
	var taskName string
	var taskID uint16
	var taskList []*task.Task

	var taskService = service.NewTaskService()

	for {
		fmt.Println("Commands:")
		fmt.Println("1. add")
		fmt.Println("2. list")
		fmt.Println("3. complete")

		fmt.Scanf("%d\n", &choice)

		switch choice {
		case 1:
			fmt.Printf("Task name: ")
			fmt.Scanf("%s\n", &taskName)
			taskService.Add(taskName)
		case 2:
			taskList = taskService.List()
			fmt.Println("ID   Done    Created     Name")
			for _, task := range taskList {
				fmt.Printf("%d   %t   %s   %s\n", task.GetID(), task.GetDone(), task.GetCreatedAt().Format("2006-01-02 15:04:05"), task.GetName())
			}
		case 3:
			fmt.Printf("Task ID: ")
			fmt.Scanf("%d\n", &taskID)
			if err := taskService.Complete(taskID); err != nil {
				fmt.Println(err)
			}
		}
	}
}
