package workerspool

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type Job = func() (*data.Order, error)

type Pool struct {
	jobs    chan Job
	results chan *data.Order

	log logger.BaseLogger
}

func Create(count int, log logger.BaseLogger) *Pool {
	pool := &Pool{jobs: make(chan func() (*data.Order, error), count), results: make(chan *data.Order, count), log: log}
	pool.createWorkers(count)
	return pool
}

func (p *Pool) AddJob(job Job) {
	p.log.Info("[accrual:WorkersPool:AddJob] new job incoming.")
	p.jobs <- job
	p.log.Info("[accrual:WorkersPool:AddJob] new job added.")
}

func (p *Pool) CloseJobsChan() {
	p.log.Info("[accrual:WorkersPool:CloseJobsChan] jobs closed.")
	close(p.jobs)
}

func (p *Pool) GetResultChan() chan *data.Order {
	return p.results
}

func (p *Pool) CloseResultsChan() {
	p.log.Info("[accrual:WorkersPool:CloseResultsChan] results closed.")
	close(p.results)
}

func (p *Pool) createWorkers(count int) {
	for i := 0; i < count; i++ {
		go p.worker()
	}
}

func (p *Pool) worker() {
	//worker stops when jobs channel is closed.
	for job := range p.jobs {
		p.log.Info("[accrual:WorkersPool:worker] worker starts job from queue.")
		order, err := job()
		if err != nil {
			p.log.Info("[accrual:WorkersPool:worker] job finished with error: %v", err)
			continue
		}

		if order != nil {
			p.results <- order
			p.log.Info("[accrual:WorkersPool:worker] job result added to result chan.")
		} else {
			p.log.Info("[accrual:WorkersPool:worker] job failed order is nil.")
		}
	}
}
