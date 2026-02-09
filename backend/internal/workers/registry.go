package workers

import (
	"sync"

	"github.com/sherlockos/backend/internal/models"
)

// Registry keeps track of available worker types
type Registry struct {
	mu         sync.RWMutex
	registered map[models.JobType]bool
}

var (
	globalRegistry *Registry
	once           sync.Once
)

// GetGlobalRegistry returns the singleton registry instance
func GetGlobalRegistry() *Registry {
	once.Do(func() {
		globalRegistry = &Registry{
			registered: make(map[models.JobType]bool),
		}
	})
	return globalRegistry
}

// Register adds a job type to the registry
func (r *Registry) Register(jobType models.JobType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered[jobType] = true
}

// IsAvailable checks if a worker is registered for the given job type
func (r *Registry) IsAvailable(jobType models.JobType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registered[jobType]
}

// GetUnavailableReason returns a reason why a job type is unavailable
func GetUnavailableReason(jobType models.JobType) string {
	registry := GetGlobalRegistry()
	if registry.IsAvailable(jobType) {
		return ""
	}
	return "No worker registered for this job type"
}
