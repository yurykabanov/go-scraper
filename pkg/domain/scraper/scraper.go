package scraper

import (
	"strings"

	"github.com/yurykabanov/scraper/pkg/domain"
	"golang.org/x/net/html"
)

type scraper struct {
	taskRepository domain.TaskRepository
}

func New(taskRepository domain.TaskRepository) *scraper {
	return &scraper{
		taskRepository: taskRepository,
	}
}

func (s *scraper) Scrape(task *domain.FetchedTask) (*domain.ScrapedTask, error) {
	var data = make(map[string]interface{})
	var tasks []domain.Task

	definition, err := s.taskRepository.FindByName(task.TaskRef)
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(task.Result.Body)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	for _, action := range definition.Actions {
		dataChunk, newTasks := action.Apply(task, doc)

		if dataChunk != nil {
			data[action.Name()] = dataChunk
		}

		tasks = append(tasks, newTasks...)
	}

	return &domain.ScrapedTask{FetchedTask: *task, NewTasks: tasks, Data: data}, nil
}
