package throttle

import (
	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/types"
)

type NoLimitThrottle struct {
}

func (n *NoLimitThrottle) RegisterRequest() {}
func (n *NoLimitThrottle) WaitIfBlocked()   {}
func (n *NoLimitThrottle) WaitCadence()     {}

func (n *NoLimitThrottle) IsThrottling(statusCode int) bool {
	return false
}

func (n *NoLimitThrottle) HandleThrottle(statusCode int, dispatcher *core.Dispatcher, attempt types.Attempt) bool {
	return false
}

func (n *NoLimitThrottle) MarkRecovered() {}
