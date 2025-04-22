package types

import "sync"

type WorkerHandler interface {
	Start(int, *sync.WaitGroup, <-chan Attempt)
}
