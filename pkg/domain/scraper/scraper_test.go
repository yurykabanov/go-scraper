package scraper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yurykabanov/scraper/pkg/domain"
	"golang.org/x/net/html"
)

type taskRepository struct {
	tasks map[string]domain.TaskDefinition
}

func (r *taskRepository) FindByName(name string) (*domain.TaskDefinition, error) {
	if task, ok := r.tasks[name]; ok {
		return &task, nil
	}
	return nil, errors.New("task not found")
}

func TestScraper_Scrape_ProcessesTask(t *testing.T) {
	result := &domain.Result{
		StatusCode: 200,
		Url:        "http://domain.tld/some_url",
		Body:       "some body",
		Headers:    map[string][]string{},
	}

	tests := []struct {
		name    string
		task    domain.FetchedTask
		isValid bool
	}{
		{
			name: "valid task",
			task: domain.FetchedTask{
				Task: domain.Task{
					Url:     "http://domain.tld/some_url",
					TaskRef: "some_task_ref",
				},
				Result: result,
			},
			isValid: true,
		},
		{
			name: "invalid task",
			task: domain.FetchedTask{
				Task: domain.Task{
					Url:     "http://domain.tld/some_url",
					TaskRef: "non_existing_task_ref",
				},
				Result: result,
			},
			isValid: false,
		},
	}

	s := New(&taskRepository{tasks: map[string]domain.TaskDefinition{
		"some_task_ref": {Name: "some_task_ref", Actions: []domain.Action{
			&domain.TextAction{Selector: "some_selector", ContentName: "a_content_name"},
		}},
	}})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scraped, err := s.Scrape(&test.task)

			assert.Equal(t, scraped != nil, test.isValid)
			assert.Equal(t, err == nil, test.isValid)
		})
	}
}

type dummyAction struct {
	mock.Mock
	SomeRef string
}

func (a *dummyAction) Apply(task *domain.FetchedTask, node *html.Node) (interface{}, []domain.Task) {
	args := a.Called(task, node)
	return args.Get(0), args.Get(1).([]domain.Task)
}

func (a *dummyAction) Name() string {
	return a.SomeRef
}

func TestScraper_Scrape_CallsActions(t *testing.T) {
	action := &dummyAction{SomeRef: "another_task_ref"}

	s := New(&taskRepository{tasks: map[string]domain.TaskDefinition{
		"some_task_ref": {Name: "some_task_ref", Actions: []domain.Action{
			action,
		}},
	}})

	task := domain.FetchedTask{
		Task: domain.Task{
			Url:     "http://domain.tld/some_url",
			TaskRef: "some_task_ref",
		},
		Result: &domain.Result{
			StatusCode: 200,
			Url:        "http://domain.tld/some_url",
			Body:       "some body",
			Headers:    map[string][]string{},
		},
	}

	result := "something"
	newTasks := []domain.Task{
		{Url:"http://domain.tld/some_new_url", TaskRef: "new_task_ref"},
	}

	action.On("Apply", &task, mock.Anything).Return(result, newTasks).Once()

	scraped, _ := s.Scrape(&task)

	action.AssertExpectations(t)

	assert.Equal(t, result, scraped.Data["another_task_ref"])
	assert.Equal(t, newTasks, scraped.NewTasks)
}
