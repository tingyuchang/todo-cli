package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Task struct {
	Id       string
	Desc     string
	Status   int
	CreateAt string
	UpdateAt string
}

const (
	StatusBacklog int = iota
	StatusInProgress
	StatusReview
	StatusDone
)

var tasks []Task
var index map[string]int
var CSVFILE = "/todo.csv"

func main() {
	start := time.Now()
	defer func() {
		fmt.Println("Execution time: ", time.Since(start))
	}()

	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return
	}
	currentPath := filepath.Dir(ex)
	CSVFILE = currentPath + CSVFILE

	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage: insert | search | update | delete | list <argument>")
		return
	}

	_, err = os.Stat(CSVFILE)
	if err != nil {
		// file not exist, create new one
		fmt.Println("data is not exist, creating...")
		f, err := os.Create(CSVFILE)
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Close()
	}

	fileInfo, err := os.Stat(CSVFILE)
	mode := fileInfo.Mode()
	if !mode.IsRegular() {
		fmt.Println(CSVFILE, "not a regular file")
		return
	}

	err = readCSVFile(CSVFILE)
	if err != nil {
		fmt.Println(err)
		return
	}

	createIndex()

	switch args[1] {
	case "insert":
		if len(args) < 3 {
			fmt.Println("Usage: insert description")
			return
		}

		// collect task description with space like: go home
		var desc string
		for i, v := range args[2:] {
			if i > 0 {
				desc += " "
			}
			desc += v
		}

		temp := initT(desc)
		if temp != nil {
			err := insert(temp)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	case "search":
		if len(args) != 4 {
			fmt.Println("Usage: search id/status <argument>")
			return
		}
		// id or status
		action := args[2]
		if action == "id" {
			t := search(args[3])
			if t == nil {
				fmt.Printf("task %s is not found.\n", args[3])
				return
			}
			fmt.Println(*t)
		}

		if action == "status" {
			t := filterStatus(args[3])
			if len(t) == 0 {
				fmt.Printf("No status %s tasks.\n", args[3])
				return
			}
			for _, v := range t {
				fmt.Println(*v)
			}
		}

	case "update":
		if len(args) != 4 {
			fmt.Println("Usage: update id status")
			return
		}

		t, err := update(args[2], args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(*t)

	case "delete":
		if len(args) != 3 {
			fmt.Println("Usage: delete id")
			return
		}

		err := deleteTask(args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Task %s has been deleted", args[2])
	case "list":
		list()
	default:
		fmt.Println("Unsupported instruction")
	}
}

func readCSVFile(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}
	for _, line := range lines {
		status, _ := strconv.Atoi(line[2])
		temp := Task{
			Id:       line[0],
			Desc:     line[1],
			Status:   status,
			CreateAt: line[3],
			UpdateAt: line[4],
		}
		tasks = append(tasks, temp)
	}

	return nil
}

func saveCSVFile(filePath string) error {
	csvFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	csvWriter := csv.NewWriter(csvFile)
	for _, v := range tasks {
		temp := []string{v.Id, v.Desc, strconv.Itoa(v.Status), v.CreateAt, v.UpdateAt}
		_ = csvWriter.Write(temp)
	}

	csvWriter.Flush()

	return nil
}

func createIndex() {
	index = make(map[string]int)
	for i, v := range tasks {
		index[v.Id] = i
	}
}

func initT(description string) *Task {
	if description == "" {
		return nil
	}
	return &Task{
		Id:       getId(),
		Desc:     description,
		Status:   StatusBacklog,
		CreateAt: strconv.FormatInt(time.Now().Unix(), 10),
		UpdateAt: strconv.FormatInt(time.Now().Unix(), 10),
	}
}

func insert(task *Task) error {
	_, ok := index[(*task).Id]
	if ok {
		return fmt.Errorf("%s already exists", (*task).Id)
	}
	tasks = append(tasks, *task)
	index[(*task).Id] = len(tasks) - 1

	err := saveCSVFile(CSVFILE)
	if err != nil {
		return err
	}
	return nil
}

// getId generate random string
func getId() string {
	if len(tasks) == 0 {
		return "1"
	}
	lastId, _ := strconv.Atoi(tasks[len(tasks)-1].Id)
	return strconv.Itoa(lastId + 1)
}

func list() {
	for _, v := range tasks {
		fmt.Println(v)
	}
}

func search(key string) *Task {
	i, ok := index[key]
	if !ok {
		return nil
	}
	return &tasks[i]
}

func filterStatus(key string) []*Task {
	result := make([]*Task, 0)

	for _, v := range tasks {
		if strconv.Itoa(v.Status) == key {
			result = append(result, &v)
		}
	}
	return result
}

func update(id, status string) (*Task, error) {
	t := search(id)
	if t == nil {
		return nil, fmt.Errorf("%s is not found\n", id)
	}
	s, _ := strconv.Atoi(status)

	if s > StatusDone {
		return nil, fmt.Errorf("%s is not valid status\n", status)
	}
	(*t).Status = s
	_ = saveCSVFile(CSVFILE)
	return t, nil
}

// deleteTask will delete task from tasks and index
func deleteTask(id string) error {
	i, ok := index[id]
	if !ok {
		return fmt.Errorf("Task %s is not found.\n", id)
	}
	// can discuss other way to remove data from slice
	tasks = append(tasks[:i], tasks[i+1:]...)
	err := saveCSVFile(CSVFILE)
	if err != nil {
		return err
	}
	// remove from index
	delete(index, id)
	return nil
}