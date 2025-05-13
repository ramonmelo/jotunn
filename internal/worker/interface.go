package worker

import (
	"sync"

	"github.com/LinharesAron/jotunn/internal/types"
)

type Worker interface {
	Start(int, *sync.WaitGroup, <-chan types.Attempt, func(types.Attempt) error)
}
