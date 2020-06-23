package pipeline

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/yurykabanov/scraper/pkg/domain"
)

type ScraperStage struct {
	workers int
	buffer  int
	input   <-chan *domain.FetchedTask
	output  chan *domain.ScrapedTask
	scraper domain.Scraper
}

type ScraperStageOption func(f *ScraperStage)

func WithScraperWorkers(c int) ScraperStageOption {
	return func(f *ScraperStage) {
		f.workers = c
	}
}

func WithScraperBuffer(c int) ScraperStageOption {
	return func(f *ScraperStage) {
		f.buffer = c
	}
}

func Scraper(
	input <-chan *domain.FetchedTask,
	scraper domain.Scraper,
	opts ...ScraperStageOption,
) *ScraperStage {
	s := &ScraperStage{
		workers: 1,
		buffer:  16,
		input:   input,
		scraper: scraper,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.output = make(chan *domain.ScrapedTask, s.buffer)

	return s
}

func (s *ScraperStage) RunStage() {
	wg := sync.WaitGroup{}
	wg.Add(s.workers)

	for i := 0; i < s.workers; i++ {
		go func(i int) {
			for task := range s.input {
				result, err := s.scraper.Scrape(task)
				if err != nil {
					// TODO: scraper failure means something really bad is happening? shouldn't this be a fatal?
					log.WithError(err).WithFields(log.Fields{
						"task_id": task.Identity(),
						"task_url": task.Url,
						"worker": i,
					}).Error("Scraper failure")
				}

				s.output <- result
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	close(s.output)
}

func (s *ScraperStage) Output() <-chan *domain.ScrapedTask {
	return s.output
}
