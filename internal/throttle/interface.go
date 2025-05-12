package throttle

type Throttler interface {
	RegisterRequest()
	WaitIfBlocked()
	WaitCadence()
	IsThrottling(statusCode int) bool
	Trigger()
	MarkRecovered()
}
