package pipeline

import (
	"github.com/yurykabanov/scraper/pkg/domain"
)

type TaskCallback func(*domain.Task)
type FailedTaskCallback func(*domain.FailedTask)
type FetchedTaskCallback func(*domain.FetchedTask)
type ScrapedTaskCallback func(*domain.ScrapedTask)

type SinkTaskCb struct {
	input <-chan *domain.Task
	cb    TaskCallback
}

func SinkTaskFunc(input <-chan *domain.Task, cb TaskCallback) *SinkTaskCb {
	return &SinkTaskCb{
		input: input,
		cb:    cb,
	}
}

func (s *SinkTaskCb) RunStage() {
	for t := range s.input {
		s.cb(t)
	}
}

type SinkFailedTaskCb struct {
	input <-chan *domain.FailedTask
	cb    FailedTaskCallback
}

func SinkFailedTaskFunc(input <-chan *domain.FailedTask, cb FailedTaskCallback) *SinkFailedTaskCb {
	return &SinkFailedTaskCb{
		input: input,
		cb:    cb,
	}
}

func (s *SinkFailedTaskCb) RunStage() {
	for t := range s.input {
		s.cb(t)
	}
}

type SinkFetchedTaskCb struct {
	input <-chan *domain.FetchedTask
	cb    FetchedTaskCallback
}

func SinkFetchedTaskFunc(input <-chan *domain.FetchedTask, cb FetchedTaskCallback) *SinkFetchedTaskCb {
	return &SinkFetchedTaskCb{
		input: input,
		cb:    cb,
	}
}

func (s *SinkFetchedTaskCb) RunStage() {
	for t := range s.input {
		s.cb(t)
	}
}

type SinkScrapedTaskCb struct {
	input <-chan *domain.ScrapedTask
	cb    ScrapedTaskCallback
}

func SinkScrapedTaskFunc(input <-chan *domain.ScrapedTask, cb ScrapedTaskCallback) *SinkScrapedTaskCb {
	return &SinkScrapedTaskCb{
		input: input,
		cb:    cb,
	}
}

func (s *SinkScrapedTaskCb) RunStage() {
	for t := range s.input {
		s.cb(t)
	}
}
