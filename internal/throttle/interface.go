package throttle

type Throttler interface {
	Wait()
	Trigger()
	MarkRecovered()
}
