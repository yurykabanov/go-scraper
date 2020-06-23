package pipeline

import (
	"sync"

	"github.com/yurykabanov/scraper/pkg/domain"
)

func TaskMerger(cs ...<-chan *domain.Task) <-chan *domain.Task {
	var wg sync.WaitGroup
	out := make(chan *domain.Task, 32)

	wg.Add(len(cs))

	for _, c := range cs {
		go func(ch <-chan *domain.Task) {
			for resp := range ch {
				out <- resp
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func FetchedTaskMerger(cs ...<-chan *domain.FetchedTask) <-chan *domain.FetchedTask {
	var wg sync.WaitGroup
	out := make(chan *domain.FetchedTask, 32)

	wg.Add(len(cs))

	for _, c := range cs {
		go func(ch <-chan *domain.FetchedTask) {
			for resp := range ch {
				out <- resp
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func FailedTaskMerger(cs ...<-chan *domain.FailedTask) <-chan *domain.FailedTask {
	var wg sync.WaitGroup
	out := make(chan *domain.FailedTask, 32)

	wg.Add(len(cs))

	for _, c := range cs {
		go func(ch <-chan *domain.FailedTask) {
			for resp := range ch {
				out <- resp
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
