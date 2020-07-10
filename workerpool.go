package workerpool

import (
	"fmt"
	"sync"

	"github.com/gosuri/uiprogress"
)

type Pool struct {
	jobs []func()

	concurrency int
	jobChan     chan func()
	wg          sync.WaitGroup

	bar *uiprogress.Bar
}

func New(concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		jobChan:     make(chan func()),
	}
}

func (p *Pool) Add(job func()) {
	p.jobs = append(p.jobs, job)
}

func (p *Pool) Run(progress bool) {
	if progress {
		uiprogress.Start()
		p.bar = uiprogress.AddBar(len(p.jobs))
		p.bar.AppendCompleted()
		p.bar.PrependElapsed()
		p.bar.PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Task (%d/%d)", b.Current(), len(p.jobs))
		})
	}

	for i := 0; i < p.concurrency; i++ {
		go func() {
			p.work(i)
		}()
	}

	p.wg.Add(len(p.jobs))
	for _, job := range p.jobs {
		p.jobChan <- job
	}

	close(p.jobChan)

	p.wg.Wait()

	if progress {
		uiprogress.Stop()
	}
}

func (p *Pool) work(i int) {
	for job := range p.jobChan {
		job()
		p.wg.Done()
		if p.bar != nil {
			p.bar.Incr()
		}
	}
}
