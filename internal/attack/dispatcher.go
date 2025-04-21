package attack

import (
	"math/rand"
	"sync"

	"github.com/LinharesAron/jotunn/internal/config"
)

type Dispatcher struct {
	workers  chan Attempt
	workerWg *sync.WaitGroup

	retries   chan Attempt
	retriesWg *sync.WaitGroup
}

func NewDispatcher(numWorkers int, bufferSize int) *Dispatcher {
	d := &Dispatcher{
		workers:  make(chan Attempt, bufferSize),
		workerWg: &sync.WaitGroup{},

		retries:   make(chan Attempt, bufferSize),
		retriesWg: &sync.WaitGroup{},
	}
	return d
}

func (d *Dispatcher) DistributeToWorkers(users []string, passwords []string) {
	numberDispatchers := 10
	userCount := len(users)

	var distWg sync.WaitGroup
	distWg.Add(numberDispatchers)

	for i := range numberDispatchers {
		go func(offset int) {
			defer distWg.Done()
			for j := offset; j < userCount; j += numberDispatchers {
				user := users[j]
				for _, pass := range passwords {
					d.Dispatch(Attempt{Username: user, Password: pass})
				}
			}
		}(i)
	}

	distWg.Wait()
}

func (d *Dispatcher) StartWorkersHandler(cfg *config.AttackConfig, limiter *RateLimitManager) {
	for i := range cfg.Threads {
		d.workerWg.Add(1)
		go Worker(i, cfg, d.workers, d, d.workerWg, limiter)
	}
}

func (d *Dispatcher) StartRetryHandler(cfg *config.AttackConfig, limiter *RateLimitManager) {
	go func() {
		for attempt := range d.retries {
			d.retriesWg.Add(1)
			go func(at Attempt) {
				defer d.retriesWg.Done()
				ch := make(chan Attempt, 1)
				ch <- at
				close(ch)
				Worker(rand.Intn(9999), cfg, ch, d, d.workerWg, limiter)
			}(attempt)
		}
	}()
}

func (d *Dispatcher) WaitWorkers() {
	d.workerWg.Wait()
}

func (d *Dispatcher) WaitRetries() {
	d.retriesWg.Wait()
}

func (d *Dispatcher) Dispatch(attempt Attempt) {
	d.workers <- attempt
}

func (d *Dispatcher) Retry(attempt Attempt) {
	d.retries <- attempt
}

func (d *Dispatcher) CloseWorkers() {
	close(d.workers)
}

func (d *Dispatcher) CloseRetries() {
	close(d.retries)
}
