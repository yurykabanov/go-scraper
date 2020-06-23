package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Task struct {
	Url     string `yaml:"url"`
	TaskRef string `yaml:"task_ref"`
}

func (t *Task) String() string {
	return fmt.Sprintf("{TaskRef: '%s', SourceLink: '%s'}", t.TaskRef, t.Url)
}

func (t *Task) Identity() string {
	return t.TaskRef + "/" + t.Hash()
}

func (t *Task) Hash() string {
	h := sha256.New()
	h.Write([]byte(t.Url))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

type FailedTask struct {
	Task
	Error error
}

func (t *FailedTask) String() string {
	return fmt.Sprintf("{TaskRef: '%s', SourceLink: '%s', Error: '%s'}", t.TaskRef, t.Url, t.Error.Error())
}

type FetchedTask struct {
	Task
	Result *Result
}

func (t *FetchedTask) String() string {
	return fmt.Sprintf("{TaskRef: '%s', SourceLink: '%s', ResultCode: %d, Body: <%d bytes>}",
		t.TaskRef, t.Url, t.Result.StatusCode, len(t.Result.Body))
}

type ScrapedTask struct {
	FetchedTask
	Data     map[string]interface{}
	NewTasks []Task
}

func (t *ScrapedTask) String() string {
	return fmt.Sprintf("{TaskRef: '%s', SourceLink: '%s', Data: <%d fields>, NewTasks: %d}",
		t.TaskRef, t.Url, len(t.Data), len(t.NewTasks))
}

type TaskRepository interface {
	FindByName(name string) (*TaskDefinition, error)
}

type mapTaskRepository struct {
	definitions map[string]TaskDefinition
}

func NewMapTaskRepository(definitions map[string]TaskDefinition) *mapTaskRepository {
	return &mapTaskRepository{
		definitions: definitions,
	}
}

func (r *mapTaskRepository) FindByName(name string) (*TaskDefinition, error) {
	val, _ := r.definitions[name]
	return &val, nil
}
