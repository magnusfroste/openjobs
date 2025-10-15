package models

import "time"

// PluginConnector defines the interface that all plugin connectors must implement
type PluginConnector interface {
	GetID() string
	GetName() string
	FetchJobs() ([]JobPost, error)
	SyncJobs() error
}

// PluginConfig represents plugin configuration stored in database
type PluginConfig struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Type      string                 `json:"type" db:"type"`
	Enabled   bool                   `json:"enabled" db:"enabled"`
	Config    map[string]interface{} `json:"config" db:"config"`
	LastRun   time.Time              `json:"last_run" db:"last_run"`
	NextRun   time.Time              `json:"next_run" db:"next_run"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// PluginRegistry holds all active plugin connectors
type PluginRegistry struct {
	connectors map[string]PluginConnector
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		connectors: make(map[string]PluginConnector),
	}
}

// Register adds a plugin connector to the registry
func (pr *PluginRegistry) Register(connector PluginConnector) {
	pr.connectors[connector.GetID()] = connector
}

// Unregister removes a plugin connector from the registry
func (pr *PluginRegistry) Unregister(id string) {
	delete(pr.connectors, id)
}

// GetConnector retrieves a registered connector by ID
func (pr *PluginRegistry) GetConnector(id string) (PluginConnector, bool) {
	connector, exists := pr.connectors[id]
	return connector, exists
}

// GetAllConnectors returns all registered connectors
func (pr *PluginRegistry) GetAllConnectors() map[string]PluginConnector {
	return pr.connectors
}

// GetEnabledConnectors returns only enabled connectors
func (pr *PluginRegistry) GetEnabledConnectors() []PluginConnector {
	var enabled []PluginConnector
	for _, connector := range pr.connectors {
		// Assume all registered connectors are enabled for now
		// In future, this could check database config
		enabled = append(enabled, connector)
	}
	return enabled
}
