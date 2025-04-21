package attack

import (
	"sync"

	"github.com/LinharesAron/jotunn/internal/config"
)

type Dispatcher struct {
	attempts        chan Attempt
	workersAttempts []chan Attempt
	retries         chan Attempt
	workersRetries  []chan Attempt
	attemptWg       *sync.WaitGroup
	retriesWg       *sync.WaitGroup
	done            chan struct{}
}

func NewDispatcher(numWorkers int, numRetries int, bufferSize int) *Dispatcher {
	d := &Dispatcher{
		attempts:        make(chan Attempt, bufferSize),
		workersAttempts: make([]chan Attempt, numWorkers),

		retries:        make(chan Attempt, bufferSize),
		workersRetries: make([]chan Attempt, numRetries),
		attemptWg:      &sync.WaitGroup{},
		retriesWg:      &sync.WaitGroup{},
		done:           make(chan struct{}),
	}

	for i := range numWorkers {
		d.workersAttempts[i] = make(chan Attempt, bufferSize)
	}

	for i := range numRetries {
		d.workersRetries[i] = make(chan Attempt, bufferSize)
	}

	return d
}

func distributeTo(workers []chan Attempt, channel <-chan Attempt) {
	idx := 0
	for attempt := range channel {
		workers[idx] <- attempt
		idx = (idx + 1) % len(workers)
	}

	for _, ch := range workers {
		close(ch)
	}
}

func (d *Dispatcher) Start(cfg *config.AttackConfig, limiter *RateLimitManager) {
	go func() {
		distributeTo(d.workersAttempts, d.attempts)
		distributeTo(d.workersRetries, d.retries)
	}()
}

func (d *Dispatcher) StartWorkers(cfg *config.AttackConfig, limiter *RateLimitManager) {
	for i := range cfg.Threads {
		d.attemptWg.Add(1)
		go Worker(i, cfg, d.WorkerAttemptQueue(i), d, d.attemptWg, limiter)
	}

	for i := range cfg.ThreadsRetry {
		d.retriesWg.Add(1)
		go Worker(i+cfg.Threads, cfg, d.WorkerRetriesQueue(i), d, d.retriesWg, limiter)
	}
}

func (d *Dispatcher) WaitAttemps() {
	d.attemptWg.Wait()
}

func (d *Dispatcher) WaitRetries() {
	d.retriesWg.Wait()
}

func (d *Dispatcher) Dispatch(attempt Attempt) {
	d.attempts <- attempt
}

func (d *Dispatcher) Retry(attempt Attempt) {
	d.retries <- attempt
}

func (d *Dispatcher) WorkerAttemptQueue(index int) <-chan Attempt {
	return d.workersAttempts[index]
}

func (d *Dispatcher) WorkerRetriesQueue(index int) <-chan Attempt {
	return d.workersRetries[index]
}

func (d *Dispatcher) CloseAttempts() {
	close(d.attempts)
}

func (d *Dispatcher) CloseRetries() {
	close(d.retries)
}
