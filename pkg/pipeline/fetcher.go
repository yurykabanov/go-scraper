package pipeline

import (
	"sync"

	"github.com/yurykabanov/scraper/pkg/domain"
)

type FetcherStage struct {
	workers int

	resultsBuffer int
	errorsBuffer  int

	source  <-chan *domain.Task
	results chan *domain.FetchedTask
	errors  chan *domain.FailedTask

	fetcher domain.Fetcher
}

type FetcherOption func(f *FetcherStage)

func WithFetcherWorkers(c int) FetcherOption {
	return func(f *FetcherStage) {
		f.workers = c
	}
}

func WithFetcherResultsBuffer(c int) FetcherOption {
	return func(f *FetcherStage) {
		f.resultsBuffer = c
	}
}

func WithFetcherErrorsBuffer(c int) FetcherOption {
	return func(f *FetcherStage) {
		f.errorsBuffer = c
	}
}

func Fetcher(
	source <-chan *domain.Task,
	fetcher domain.Fetcher,
	opts ...FetcherOption,
) *FetcherStage {
	f := &FetcherStage{
		workers: 1,

		resultsBuffer: 256,
		errorsBuffer:  64,

		source:  source,
		fetcher: fetcher,
	}

	for _, opt := range opts {
		opt(f)
	}

	f.results = make(chan *domain.FetchedTask, f.resultsBuffer)
	f.errors = make(chan *domain.FailedTask, f.errorsBuffer)

	return f
}

func (s *FetcherStage) RunStage() {
	wg := sync.WaitGroup{}
	wg.Add(s.workers)

	for i := 0; i < s.workers; i++ {
		go s.runWorker(i, &wg)
	}

	wg.Wait()

	close(s.results)
	close(s.errors)
}

func (s *FetcherStage) runWorker(i int, wg *sync.WaitGroup) {
	for task := range s.source {
		res, err := s.fetcher.Fetch(task)
		if err != nil {
			s.errors <- &domain.FailedTask{Task: *task, Error: err}
			continue
		}
		s.results <- &domain.FetchedTask{Task: *task, Result: res}
	}
	wg.Done()
}

func (s *FetcherStage) Results() <-chan *domain.FetchedTask {
	return s.results
}

func (s *FetcherStage) Errors() <-chan *domain.FailedTask {
	return s.errors
}
