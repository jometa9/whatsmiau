package instances

import (
	"sync"

	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/models"
	"golang.org/x/net/context"
)

// These verify if MemoryInstance follows instances interface pattern
var _ interfaces.InstanceRepository = (*MemoryInstance)(nil)

type MemoryInstance struct {
	mu        sync.RWMutex
	instances map[string]*models.Instance
}

func NewMemory() *MemoryInstance {
	return &MemoryInstance{
		instances: make(map[string]*models.Instance),
	}
}

func (s *MemoryInstance) Create(ctx context.Context, instance *models.Instance) error {
	if instance.ID == "" {
		return ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.instances[instance.ID]; exists {
		return ErrorAlreadyExists
	}

	// Create a copy to avoid external modifications
	instCopy := *instance
	s.instances[instance.ID] = &instCopy

	return nil
}

func (s *MemoryInstance) Update(ctx context.Context, id string, toUpdate *models.Instance) (*models.Instance, error) {
	if id == "" {
		return nil, ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldInstance, exists := s.instances[id]
	if !exists {
		return nil, ErrorNotFound
	}

	// Update fields
	if len(toUpdate.RemoteJID) > 0 {
		oldInstance.RemoteJID = toUpdate.RemoteJID
	}
	if toUpdate.Webhook.Url != "" {
		oldInstance.Webhook.Url = toUpdate.Webhook.Url
	}
	if toUpdate.Webhook.ByEvents != nil {
		oldInstance.Webhook.ByEvents = toUpdate.Webhook.ByEvents
	}
	if toUpdate.Webhook.Base64 != nil {
		oldInstance.Webhook.Base64 = toUpdate.Webhook.Base64
	}
	if toUpdate.Webhook.Headers != nil {
		if oldInstance.Webhook.Headers == nil {
			oldInstance.Webhook.Headers = make(map[string]string)
		}
		for k, v := range toUpdate.Webhook.Headers {
			oldInstance.Webhook.Headers[k] = v
		}
	}
	if toUpdate.Webhook.Events != nil && len(toUpdate.Webhook.Events) > 0 {
		oldInstance.Webhook.Events = toUpdate.Webhook.Events
	}

	// Return a copy
	result := *oldInstance
	return &result, nil
}

func (s *MemoryInstance) List(ctx context.Context, id string) ([]models.Instance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id != "" {
		instance, exists := s.instances[id]
		if !exists {
			return []models.Instance{}, nil
		}
		// Return a copy
		return []models.Instance{*instance}, nil
	}

	// Return all instances
	instances := make([]models.Instance, 0, len(s.instances))
	for _, inst := range s.instances {
		instances = append(instances, *inst)
	}

	return instances, nil
}

func (s *MemoryInstance) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInstanceIDEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.instances[id]; !exists {
		return ErrorNotFound
	}

	delete(s.instances, id)
	return nil
}

