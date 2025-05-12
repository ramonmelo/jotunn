package throttle

type NoLimitThrottle struct {
}

func (n *NoLimitThrottle) RegisterRequest() {}
func (n *NoLimitThrottle) WaitIfBlocked()   {}
func (n *NoLimitThrottle) WaitCadence()     {}

func (n *NoLimitThrottle) IsThrottling(statusCode int) bool {
	return false
}

func (n *NoLimitThrottle) Trigger()       {}
func (n *NoLimitThrottle) MarkRecovered() {}
