package list

import (
	"errors"
	"fmt"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/file"
	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks in to do list",
		RunE: func(cmd *cobra.Command, args []string) error {
			taskService := task.NewTaskService()
			fileSystem := file.NewFileService(taskService)

			taskFile, err := cmd.Flags().GetString("path")
			if err != nil {
				return err
			}
			listErr := fileSystem.ReadTasksFromFile(taskFile)

			if len(listErr) > 0 {
				for _, err := range listErr {
					fmt.Println(err)
				}
				return errors.New("All the errors are listed above")
			}

			showStatus, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}
			taskService.StdOutPrint(showStatus)
			return nil
		},
	}
	cmd.Flags().String("path", "", "Path of the task file (required)")
	cmd.MarkFlagRequired("path")
	cmd.Flags().BoolP("all", "a", false, "List all tasks with status")
	return cmd
}
