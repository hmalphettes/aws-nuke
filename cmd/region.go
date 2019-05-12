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
	}
}

func (region *Region) Session(resourceType string) (*session.Session, error) {
	if region.lock == nil {
		region.lock = &sync.RWMutex{}
	}

	// Need to read
	region.lock.RLock()
	defer region.lock.RUnlock()
	sess := region.cache[resourceType]
	if sess != nil {
		return sess, nil
	}
	region.lock.RUnlock()

	// Need to write:
	region.lock.Lock()
	defer region.lock.Unlock()
	sess, err := region.NewSession(region.Name, resourceType)
	if err != nil {
		return nil, err
	}
	region.cache[resourceType] = sess
	return sess, nil
}
