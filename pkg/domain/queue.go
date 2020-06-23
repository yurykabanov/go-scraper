package domain

import (
	"log"
	"sync"
)

type TaskQueue interface {
	Push(Task)
	PushBatch([]Task)
	Chan() <-chan *Task
	Close()
}

type MemoryTaskQueue struct {
	mu *sync.Mutex
	cond  *sync.Cond
	output chan *Task
	notify chan struct{}
	items  []*Task

	head int
	tail int
}

func NewMemoryTaskQueue() *MemoryTaskQueue {
	mu := &sync.Mutex{}

	q := &MemoryTaskQueue{
		mu:     mu,
		cond:   sync.NewCond(mu),
		notify: make(chan struct{}, 1),
		output: make(chan *Task, 1),
	}

	go q.run()
	<-q.notify

	return q
}

func (q *MemoryTaskQueue) run() {
	log.Print(123)
	close(q.notify)

	for {
		//q.mu.Lock()

		log.Println("RUN >>> before lock")
		q.cond.L.Lock()
		log.Println("RUN >>> after lock")

		log.Println("RUN >>> before wait")
		q.cond.Wait()
		log.Println("RUN >>> after wait")

		item := q.items[q.head]
		q.items = q.items[q.head+1:]
		q.head++

		log.Println("RUN >>> before unlock")
		q.cond.L.Unlock()
		log.Println("RUN >>> after unlock")

		//q.mu.Unlock()

		q.output <- item
	}
}

func (q *MemoryTaskQueue) Push(task *Task) {
	log.Println("PUSH >>> before lock")
	q.cond.L.Lock()
	log.Println("PUSH >>> after lock")
	//q.mu.Lock()

	q.tail++
	q.items = append(q.items, task)

	q.cond.Signal()

	//select {
	//case q.notify <- struct{}{}:
	//default:
	//	log.Println("not notified")
	//}

	//q.mu.Unlock()

	log.Println("PUSH >>> before unlock")
	q.cond.L.Unlock()
	log.Println("PUSH >>> after unlock")
}


func (q *MemoryTaskQueue) Chan() <-chan *Task {
	return q.output
}

func (q *MemoryTaskQueue) Close() {
	close(q.output)

}


//func (q *MemoryTaskQueue) PushBatch(tasks []*Task) {
//	q.mu.Lock()
//
//	q.tail += len(tasks)
//	q.items = append(q.items, tasks...)
//	select {
//	case q.notify <- struct{}{}:
//	default:
//	}
//
//	q.mu.Unlock()
//}
