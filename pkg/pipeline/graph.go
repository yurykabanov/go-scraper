package pipeline

import (
	"sync"
)

type Stage interface {
	RunStage()
}

type Graph struct {
	stages []Stage
}

func (g *Graph) RunStage() {
	wg := sync.WaitGroup{}
	wg.Add(len(g.stages))

	for _, s := range g.stages {
		go func(s Stage) {
			s.RunStage()
			wg.Done()
		}(s)
	}

	wg.Wait()
}

func NewGraph(stages ...Stage) *Graph {
	return &Graph{
		stages: stages,
	}
}
