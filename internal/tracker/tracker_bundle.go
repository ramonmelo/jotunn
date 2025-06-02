package tracker

import "sync"

type TrackerBundle struct {
	Attempts   Tracker
	Credential Tracker
	basePath   string
}

var (
	bundleInstance *TrackerBundle
	once           sync.Once
)

func InitTracker(basePath string) *TrackerBundle {
	once.Do(func() {
		bundleInstance = newTrackerBundle(basePath)
	})
	return bundleInstance
}

func Get() *TrackerBundle {
	if bundleInstance == nil {
		panic("TrackerBundle has not been initialized. Call InitTracker() first.")
	}
	return bundleInstance
}

func newTrackerBundle(basePath string) *TrackerBundle {
	return &TrackerBundle{
		basePath: basePath,
	}
}

func (b *TrackerBundle) StartAttempts(filename string) *TrackerBundle {
	b.Attempts = newAttemptTracker(b.basePath, filename)
	return b
}

func (b *TrackerBundle) StartCredential() *TrackerBundle {
	b.Credential = newCredentialTracker(b.basePath)
	return b
}

func (b *TrackerBundle) CloseAll() {
	if b.Attempts != nil {
		b.Attempts.Close()
	}
	if b.Credential != nil {
		b.Credential.Close()
	}
}
