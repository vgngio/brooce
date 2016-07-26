package task

import (
	"encoding/json"
	"log"
	"strings"
)

type Task struct {
	Id      string   `json:"id"`
	Command []string `json:"command"`
}

func NewFromJson(str string) (task *Task, err error) {
	task = &Task{}
	err = json.Unmarshal([]byte(str), task)
	return
}

func (task *Task) Json() string {
	bytes, err := json.Marshal(task)
	if err != nil {
		log.Fatalln(err)
	}

	return string(bytes)
}

func (task *Task) FullCommand() string {
	if task.Command == nil {
		task.Command = []string{}
	}

	return strings.Join(task.Command, " ")
}
