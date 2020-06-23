package pipeline

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/yurykabanov/scraper/pkg/domain"
)

type SupervisorStage struct {
	input   <-chan *domain.Task
	Scraped <-chan *domain.ScrapedTask
	Failed  <-chan *domain.FailedTask
	output  chan *domain.Task

	mu      *sync.Mutex
	tasks   map[string]bool
	counter int64
}

func Supervisor(
	input <-chan *domain.Task,
) *SupervisorStage {
	return &SupervisorStage{
		input:  input,
		output: make(chan *domain.Task, 100000),

		mu:      &sync.Mutex{},
		tasks:   make(map[string]bool),
		counter: 0,
	}
}

func (s *SupervisorStage) Output() <-chan *domain.Task {
	return s.output
}

func (s *SupervisorStage) RunStage() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for task := range s.Scraped {
			s.mu.Lock()

			s.counter -= 1
			s.tasks[task.Result.Hash()] = true

			duplicates := 0

			for i := range task.NewTasks {
				id := task.NewTasks[i].Identity()

				if _, ok := s.tasks[id]; ok {
					duplicates += 1
					continue
				}

				s.tasks[id] = false
				s.counter += 1

				// TODO: this part could be outside of this critical section
				s.output <- &task.NewTasks[i]
			}

			s.mu.Unlock()

			newTasks := len(task.NewTasks)
			log.WithFields(log.Fields{
				"total_new_tasks": newTasks,
				"unique_new_tasks": newTasks - duplicates,
				"duplicate_tasks": duplicates,
			}).Debugf("Supervisor: task scraped successfully %s", task.Task.String())

			if s.counter == 0 {
				break
			}
		}
		wg.Done()
	}()

	go func() {
		for task := range s.Failed {
			log.Debugf("Supervisor: task failed %s", task)

			// TODO: make limited amount of retries here
			s.output <- &task.Task

			// s.mu.Lock()
			// s.counter -= 1
			// s.tasks[task.Hash()] = true
			// s.mu.Unlock()
		}
	}()

	for task := range s.input {
		log.Debugf("Supervisor: new Task from source %s", task)

		s.mu.Lock()

		s.counter += 1
		s.tasks[task.Hash()] = false

		s.output <- task

		s.mu.Unlock()
	}

	wg.Wait()

	close(s.output)
}
