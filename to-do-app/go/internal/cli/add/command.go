package add

import (
	"errors"
	"fmt"
	"time"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/file"
	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return errors.New("Name cannot be empty")
			}

			path, _ := cmd.Flags().GetString("path")

			taskService := task.NewTaskService()
			fileSystem := file.NewFileService(taskService)

			listErr := fileSystem.ReadTasksFromFile(path)
			if len(listErr) > 0 {
				for _, err := range listErr {
					fmt.Println(err)
				}
				return errors.New("All the errors are listed above")
			}

			// Auto set id task
			id := len(taskService.List())

			// Then add task in the list
			taskService.Add(uint16(id), name, time.Now(), false)

			// And write task to file
			fileSystem.AppendTaskToFile(path, taskService.List()[len(taskService.List())-1])
			return nil
		},
	}
	cmd.Flags().String("name", "", "Name of a task (required)")
	cmd.MarkFlagRequired("name")
	cmd.Flags().String("path", "", "Path of the task file (required)")
	cmd.MarkFlagRequired("path")
	return cmd
}
