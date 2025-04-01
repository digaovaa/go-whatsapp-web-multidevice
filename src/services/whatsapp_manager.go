package services

import (
	"sync"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
)

type WhatsAppManager struct {
	instances map[uuid.UUID]*whatsmeow.Client
	mu        sync.RWMutex
}

var (
	manager *WhatsAppManager
	once    sync.Once
)

// GetWhatsAppManager returns the singleton instance of WhatsAppManager
func GetWhatsAppManager() *WhatsAppManager {
	once.Do(func() {
		manager = &WhatsAppManager{
			instances: make(map[uuid.UUID]*whatsmeow.Client),
		}
	})
	return manager
}

// GetInstance returns a WhatsApp client instance for a specific user
func (m *WhatsAppManager) GetInstance(userID uuid.UUID) (*whatsmeow.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, exists := m.instances[userID]
	return client, exists
}

// SetInstance sets a WhatsApp client instance for a specific user
func (m *WhatsAppManager) SetInstance(userID uuid.UUID, client *whatsmeow.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[userID] = client
}

// RemoveInstance removes a WhatsApp client instance for a specific user
func (m *WhatsAppManager) RemoveInstance(userID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.instances, userID)
}

// GetAllInstances returns all WhatsApp client instances
func (m *WhatsAppManager) GetAllInstances() map[uuid.UUID]*whatsmeow.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy of the map to avoid race conditions
	instances := make(map[uuid.UUID]*whatsmeow.Client)
	for k, v := range m.instances {
		instances[k] = v
	}
	return instances
}

// GetInstanceCount returns the number of active WhatsApp instances
func (m *WhatsAppManager) GetInstanceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.instances)
}

// DisconnectAll disconnects all WhatsApp instances
func (m *WhatsAppManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, client := range m.instances {
		if client != nil {
			client.Disconnect()
		}
	}

	// Clear the map
	m.instances = make(map[uuid.UUID]*whatsmeow.Client)
}
