package tracker

type Tracker interface {
	HasSeen(keys ...string) bool
	Mark(keys ...string)
	Close()
}
