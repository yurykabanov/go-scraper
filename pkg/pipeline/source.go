package pipeline

import (
	"github.com/yurykabanov/scraper/pkg/domain"
)

type SourceFromSliceStage struct {
	output chan *domain.Task
	tasks  []*domain.Task
}

func SourceFromSlice(tasks []*domain.Task) *SourceFromSliceStage {
	return &SourceFromSliceStage{
		output: make(chan *domain.Task),
		tasks:  tasks,
	}
}

func (s *SourceFromSliceStage) Output() <-chan *domain.Task {
	return s.output
}

func (s *SourceFromSliceStage) RunStage() {
	for _, task := range s.tasks {
		s.output <- task
	}
	close(s.output)
}
