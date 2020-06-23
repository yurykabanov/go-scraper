package pipeline

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/yurykabanov/scraper/pkg/domain"
)

type CacheLoaderStage struct {
	workers int
	input   <-chan *domain.Task
	hits    chan *domain.FetchedTask
	misses  chan *domain.Task
	cache   domain.Cache
}

type CacheLoaderOption func(f *CacheLoaderStage)

func WithLoaderWorkers(c int) CacheLoaderOption {
	return func(f *CacheLoaderStage) {
		f.workers = c
	}
}

func CacheLoader(
	input <-chan *domain.Task,
	cache domain.Cache,
	opts ...CacheLoaderOption,
) *CacheLoaderStage {
	s := &CacheLoaderStage{
		workers: 1,
		input:  input,
		hits:   make(chan *domain.FetchedTask, 64),
		misses: make(chan *domain.Task, 64),
		cache:  cache,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *CacheLoaderStage) Hits() <-chan *domain.FetchedTask {
	return s.hits
}

func (s *CacheLoaderStage) Misses() <-chan *domain.Task {
	return s.misses
}

func (s *CacheLoaderStage) RunStage() {
	wg := sync.WaitGroup{}
	wg.Add(s.workers)

	for i := 0; i < s.workers; i++ {
		go func(i int) {
			for task := range s.input {
				result, err := s.cache.Get(task.Hash())
				if err != nil {
					log.WithError(err).Error("Cache read failure")
				}

				if result == nil {
					s.misses <- task
					continue
				}

				s.hits <- &domain.FetchedTask{Task: *task, Result: result}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	close(s.hits)
	close(s.misses)
}

type CacheSaverStage struct {
	input  <-chan *domain.FetchedTask
	output chan *domain.FetchedTask
	cache  domain.Cache
}

func CacheSaver(input <-chan *domain.FetchedTask, cache domain.Cache) *CacheSaverStage {
	return &CacheSaverStage{
		input:  input,
		output: make(chan *domain.FetchedTask, 256),
		cache:  cache,
	}
}

func (s *CacheSaverStage) Output() <-chan *domain.FetchedTask {
	return s.output
}

func (s *CacheSaverStage) RunStage() {
	for task := range s.input {
		err := s.cache.Put(task.Hash(), task.Result)
		if err != nil {
			log.WithError(err).Error("Cache write failure")
		}
		s.output <- task
	}

	close(s.output)
}
