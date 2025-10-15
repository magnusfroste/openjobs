package main

import (
	"testing"

	"openjobs/connectors/arbetsformedlingen"
	"openjobs/connectors/eures"
	"openjobs/pkg/models"
	"openjobs/pkg/storage"
)

// TestPlugins verifies all registered plugin connectors implement the interface correctly
func TestPlugins(t *testing.T) {
	// Test plugin registry functionality
	registry := models.NewPluginRegistry()

	// Test that registry starts empty
	if len(registry.GetAllConnectors()) != 0 {
		t.Errorf("Expected empty registry, got %d connectors", len(registry.GetAllConnectors()))
	}

	// Test registering connectors
	mockStore := &storage.JobStore{}

	arbetsConnector := arbetsformedlingen.NewArbetsformedlingenConnector(mockStore)
	euresConnector := eures.NewEURESConnector(mockStore)

	registry.Register(arbetsConnector)
	registry.Register(euresConnector)

	// Verify connectors were registered
	enabled := registry.GetEnabledConnectors()
	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled connectors, got %d", len(enabled))
	}

	// Verify connector IDs are unique
	ids := make(map[string]bool)
	for _, connector := range enabled {
		if ids[connector.GetID()] {
			t.Errorf("Duplicate connector ID: %s", connector.GetID())
		}
		ids[connector.GetID()] = true
	}

	// Test retrieval by ID
	connector, exists := registry.GetConnector("arbetsformedlingen")
	if !exists {
		t.Error("Expected to find arbetformedlingen connector")
	}
	if connector.GetID() != "arbetsformedlingen" {
		t.Errorf("Expected ID 'arbetsformedlingen', got '%s'", connector.GetID())
	}

	// Test non-existent connector
	_, exists = registry.GetConnector("nonexistent")
	if exists {
		t.Error("Expected non-existent connector to return false")
	}
}

// TestPluginInterface verifies all plugins implement the interface correctly
func TestPluginInterface(t *testing.T) {
	// Create instances to verify interface implementation
	mockStore := &storage.JobStore{}

	arbetsConnector := arbetsformedlingen.NewArbetsformedlingenConnector(mockStore)
	euresConnector := eures.NewEURESConnector(mockStore)

	connectors := []models.PluginConnector{arbetsConnector, euresConnector}

	for i, connector := range connectors {
		// Test GetID returns non-empty string
		id := connector.GetID()
		if id == "" {
			t.Errorf("Plugin %d: GetID() returned empty string", i)
		}

		// Test GetName returns non-empty string
		name := connector.GetName()
		if name == "" {
			t.Errorf("Plugin %d: GetName() returned empty string", i)
		}

		// Verify methods don't panic (basic smoke test)
		// Note: Full functionality test would require mock HTTP server
		_ = connector.GetID()
		_ = connector.GetName()

		t.Logf("Plugin %d [%s]: '%s'", i, id, name)
	}
}
