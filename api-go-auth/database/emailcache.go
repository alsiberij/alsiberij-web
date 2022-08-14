package database

import (
	"sync"
	"time"
)

const (
	EmailCodeLifetime = 3 * time.Minute
	EmailCacheGCCycle = 5 * time.Minute
)

type (
	emailCode struct {
		Code      int
		ExpiresIn time.Time
	}
	EmailVerificationCache struct {
		data  map[string]emailCode
		mutex sync.RWMutex
	}
)

var (
	EmailCache = EmailVerificationCache{
		data:  make(map[string]emailCode),
		mutex: sync.RWMutex{},
	}
)

func (c *EmailVerificationCache) Save(email string, code int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[email] = emailCode{
		Code:      code,
		ExpiresIn: time.Now().Add(EmailCodeLifetime),
	}
}

func (c *EmailVerificationCache) Search(email string) (int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	v, ok := c.data[email]
	if v.ExpiresIn.Before(time.Now()) {
		v.Code, ok = 0, false
	}
	return v.Code, ok
}

func (c *EmailVerificationCache) GC() {
	ticker := time.Tick(EmailCacheGCCycle)
	for {
		select {
		case <-ticker:
			c.mutex.Lock()
			for email, code := range c.data {
				if code.ExpiresIn.Before(time.Now()) {
					delete(c.data, email)
				}
			}
			c.mutex.Unlock()
		}
	}
}
