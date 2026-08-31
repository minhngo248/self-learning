package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

func (fs *FileService) CompleteTaskInFile(filePath string, task *task.Task) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Return file cursor to the beginning of the file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	// Change file size
	if err = file.Truncate(0); err != nil {
		return err
	}

	i := 0
	var foundLine string
	for line := range bytes.Lines(content) {
		if i == int(task.GetID()) {
			foundLine = string(line)
		}
		i++
	}

	var replaceLine string
	replaceLine = strconv.Itoa(int(task.GetID())) + "," + task.GetName() + "," + task.GetCreatedAt().Format("2006-01-02 15:04:05") + "," + strconv.FormatBool(task.IsDone()) + "\r\n"

	// Replace line in content
	contentStr := string(content)
	contentStr = strings.Replace(contentStr, foundLine, replaceLine, 1)

	fileWriter := bufio.NewWriter(file)
	defer fileWriter.Flush()

	// write contentStr to file
	fileWriter.WriteString(contentStr)
	return nil
}
