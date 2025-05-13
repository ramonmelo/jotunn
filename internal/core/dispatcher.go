package core

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/LinharesAron/jotunn/internal/logger"
	"github.com/LinharesAron/jotunn/internal/types"
	"github.com/LinharesAron/jotunn/internal/worker"
)

type Dispatcher struct {
	workers  chan types.Attempt
	workerWg *sync.WaitGroup

	retries      chan types.Attempt
	retriesWg    *sync.WaitGroup
	retryTracker *RetryTracker
}

func NewDispatcher(numWorkers int, retryLimit int, bufferSize int) *Dispatcher {
	d := &Dispatcher{
		workers:  make(chan types.Attempt, bufferSize),
		workerWg: &sync.WaitGroup{},

		retries:      make(chan types.Attempt, bufferSize),
		retriesWg:    &sync.WaitGroup{},
		retryTracker: NewRetryTracker(retryLimit),
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
					d.Dispatch(types.Attempt{Username: user, Password: pass})
				}
			}
		}(i)
	}

	distWg.Wait()
}

func (d *Dispatcher) StartWorkersHandler(threads int, work worker.Worker) {
	for i := range threads {
		d.workerWg.Add(1)
		go work.Start(i, d.workerWg, d.workers, d.shouldRetry)
	}
}

func (d *Dispatcher) StartRetryHandler(work worker.Worker) {
	go func() {
		for attempt := range d.retries {
			d.retriesWg.Add(1)
			go func(at types.Attempt) {
				defer d.retriesWg.Done()
				ch := make(chan types.Attempt, 1)
				ch <- at
				close(ch)

				d.workerWg.Add(1)
				work.Start(rand.Intn(9999), d.workerWg, ch, d.shouldRetry)
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

func (d *Dispatcher) Dispatch(attempt types.Attempt) {
	d.workers <- attempt
}

func (d *Dispatcher) shouldRetry(attempt types.Attempt) error {
	if d.retryTracker.ShouldRetry(attempt) {
		d.retries <- attempt
		return nil
	}
	var ErrMaxRetries = errors.New("max retry limit reached")
	return ErrMaxRetries
}

func (d *Dispatcher) CloseWorkers() {
	close(d.workers)
	logger.Info("[~] Workers closed")
}

func (d *Dispatcher) CloseRetries() {
	close(d.retries)
}
