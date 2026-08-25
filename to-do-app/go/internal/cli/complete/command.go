package complete

import (
	"errors"
	"fmt"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/file"
	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Task is done",
		RunE: func(cmd *cobra.Command, args []string) error {
			taskId, _ := cmd.Flags().GetUint16("id")

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

			task := taskService.GetTaskById(taskId)
			if task == nil {
				return fmt.Errorf("Task with id %d not found", taskId)
			}

			// complete a task
			taskService.Complete(taskId)

			// complete a task in file
			err := fileSystem.CompleteTaskInFile(path, task)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().Uint16("id", 0, "Id of a task (required)")
	cmd.MarkFlagRequired("id")
	cmd.Flags().String("path", "", "Path of the task file (required)")
	cmd.MarkFlagRequired("path")
	return cmd
}
