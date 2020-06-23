package pipeline

import (
	"github.com/yurykabanov/scraper/pkg/domain"
)

type BroadcastScrapedTaskStage struct {
	input  <-chan *domain.ScrapedTask
	output []chan *domain.ScrapedTask
	buffer int
}

type BroadcastOption func(s *BroadcastScrapedTaskStage)

func WithBroadcastBuffer(buf int) BroadcastOption {
	return func(s *BroadcastScrapedTaskStage) {
		s.buffer = buf
	}
}

func BroadcastScrapedTask(input <-chan *domain.ScrapedTask, opts ...BroadcastOption) *BroadcastScrapedTaskStage {
	s := &BroadcastScrapedTaskStage{
		input: input,
		buffer: 16,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *BroadcastScrapedTaskStage) Output() <-chan *domain.ScrapedTask {
	ch := make(chan *domain.ScrapedTask, s.buffer)

	s.output = append(s.output, ch)

	return ch
}

func (s *BroadcastScrapedTaskStage) RunStage() {
	for task := range s.input {
		for _, ch := range s.output {
			ch <- task
		}
	}

	for _, ch := range s.output {
		close(ch)
	}
}
