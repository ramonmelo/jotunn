package throttle

type NoLimitThrottle struct {
}

func (n *NoLimitThrottle) Wait()          {}
func (n *NoLimitThrottle) Trigger()       {}
func (n *NoLimitThrottle) MarkRecovered() {}
