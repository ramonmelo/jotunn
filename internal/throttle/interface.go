package throttle

import (
	"github.com/LinharesAron/jotunn/internal/core"
	"github.com/LinharesAron/jotunn/internal/types"
)

type Throttler interface {
	RegisterRequest()
	WaitIfBlocked()
	WaitCadence()
	IsThrottling(statusCode int) bool
	HandleThrottle(statusCode int, dispatcher *core.Dispatcher, attempt types.Attempt) bool
	MarkRecovered()
}
