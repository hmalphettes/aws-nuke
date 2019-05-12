package cmd

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
)

// SessionFactory support for custom endpoints
type SessionFactory func(regionName, svcType string) (*session.Session, error)

type Region struct {
	Name       string
	NewSession SessionFactory

	cache map[string]*session.Session
	lock  *sync.RWMutex
}

func NewRegion(name string, sessionFactory SessionFactory) *Region {
	return &Region{
		Name:       name,
		NewSession: sessionFactory,
		lock:       &sync.RWMutex{},
		cache:      make(map[string]*session.Session),
	}
}

func (region *Region) Session(resourceType string) (*session.Session, error) {
	if region.lock == nil {
		region.lock = &sync.RWMutex{}
	}

	// Need to read
	region.lock.RLock()
	sess := region.cache[resourceType]
	region.lock.RUnlock()
	if sess != nil {
		return sess, nil
	}

	// Need to write:
	region.lock.Lock()
	sess, err := region.NewSession(region.Name, resourceType)
	if err != nil {
		region.lock.Unlock()
		return nil, err
	}
	region.cache[resourceType] = sess
	region.lock.Unlock()
	return sess, nil
}
