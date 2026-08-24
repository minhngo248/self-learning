package file

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minhngo248/self-learning/to-do-app/go/internal/task"
)

type FileService struct {
	taskSrv *task.TaskService
}

func NewFileService(taskSrv *task.TaskService) *FileService {
	return &FileService{
		taskSrv: taskSrv,
	}
}

func (fileSrv *FileService) ReadTasksFromFile(filePath string) []error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error, %s\n", err)
	}
	defer file.Close()

	listErr := make([]error, 0)
	fileReader := bufio.NewReader(file)

	for i := 0; ; i++ {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			break
		}
		splittedLines := strings.Split(line, ",")

		idConv, err := strconv.Atoi(splittedLines[0])
		if err != nil {
			listErr = append(listErr, err)
		}
		id := uint16(idConv)
		name := splittedLines[1]

		createdAt, err := time.Parse("2006-01-02 15:04:05", splittedLines[2])
		if err != nil {
			listErr = append(listErr, err)
		}

		done, err := strconv.ParseBool(strings.TrimSpace(splittedLines[3]))
		if err != nil {
			listErr = append(listErr, err)
		}

		if len(listErr) != 0 {
			fmt.Printf("Error(s) in line %d\n", i)
			return listErr
		}

		fileSrv.taskSrv.Add(id, name, createdAt, done)
	}

	return listErr
}

func (fs *FileService) AppendTaskToFile(filePath string, task *task.Task) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error, %s\n", err)
		return
	}
	defer file.Close()

	fileWriter := bufio.NewWriter(file)
	defer fileWriter.Flush()

	// Append task to file
	fileWriter.WriteString(strconv.Itoa(int(task.GetID())))
	fileWriter.WriteString(",")
	fileWriter.WriteString(task.GetName())
	fileWriter.WriteString(",")
	fileWriter.WriteString(task.GetCreatedAt().Format("2006-01-02 15:04:05"))
	fileWriter.WriteString(",")
	fileWriter.WriteString(strconv.FormatBool(task.IsDone()))
	fileWriter.WriteString("\n")
}
