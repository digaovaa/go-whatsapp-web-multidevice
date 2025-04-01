package whatsapp

import (
	"sync"

	"go.mau.fi/whatsmeow"
)

type Instance struct {
	Client *whatsmeow.Client
	UserID int64
}

type InstanceManager struct {
	instances map[int64]*Instance
	mu        sync.RWMutex
}

var (
	manager *InstanceManager
	once    sync.Once
)

func GetInstanceManager() *InstanceManager {
	once.Do(func() {
		manager = &InstanceManager{
			instances: make(map[int64]*Instance),
		}
	})
	return manager
}

func (im *InstanceManager) AddInstance(userID int64, client *whatsmeow.Client) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.instances[userID] = &Instance{
		Client: client,
		UserID: userID,
	}
}

// GetInstance returns an instance by user ID
func (im *InstanceManager) GetInstance(userID int64) *Instance {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.instances[userID]
}

func (im *InstanceManager) RemoveInstance(userID int64) {
	im.mu.Lock()
	defer im.mu.Unlock()
	delete(im.instances, userID)
}

func (im *InstanceManager) GetAllInstances() []*Instance {
	im.mu.RLock()
	defer im.mu.RUnlock()
	instances := make([]*Instance, 0, len(im.instances))
	for _, instance := range im.instances {
		instances = append(instances, instance)
	}
	return instances
}

// GetFirstInstance returns the first available instance from the manager
func (im *InstanceManager) GetFirstInstance() *Instance {
	im.mu.RLock()
	defer im.mu.RUnlock()

	for _, instance := range im.instances {
		return instance
	}
	return nil
}
