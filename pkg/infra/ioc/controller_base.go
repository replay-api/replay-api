package ioc

import (
	"log/slog"
	container "github.com/golobby/container/v3"
)

// ControllerBase provides common functionality for controllers
// This helps standardize dependency resolution across all controllers
type ControllerBase struct {
	container Container
}

// NewControllerBase creates a new controller base with dependency resolution helpers
func NewControllerBase(container Container) *ControllerBase {
	return &ControllerBase{container: container}
}

// Resolve safely resolves a dependency from the container
// Logs errors but doesn't panic, allowing graceful degradation
func (b *ControllerBase) Resolve(target interface{}) error {
	if err := b.container.Resolve(target); err != nil {
		slog.Error("Failed to resolve dependency", "type", target, "err", err)
		return err
	}
	return nil
}

// MustResolve resolves a dependency and panics if it fails
// Use this for required dependencies that must be available
func (b *ControllerBase) MustResolve(target interface{}) {
	if err := b.container.Resolve(target); err != nil {
		slog.Error("Failed to resolve required dependency", "type", target, "err", err)
		panic(err)
	}
}

// GetContainer returns the underlying container (for advanced use cases)
func (b *ControllerBase) GetContainer() Container {
	return b.container
}

// ContainerAdapter adapts golobby/container to our Container interface
// This allows existing code to work with the new interface
type ContainerAdapter struct {
	container *container.Container
}

// NewContainerAdapter creates an adapter from golobby container
func NewContainerAdapter(c *container.Container) Container {
	return &ContainerAdapter{container: c}
}

// Resolve implements Container interface
func (a *ContainerAdapter) Resolve(target interface{}) error {
	return a.container.Resolve(target)
}

// Singleton implements Container interface
func (a *ContainerAdapter) Singleton(resolver interface{}) error {
	return a.container.Singleton(resolver)
}

// Transient implements Container interface
func (a *ContainerAdapter) Transient(resolver interface{}) error {
	return a.container.Transient(resolver)
}

// Scoped implements Container interface
// Note: golobby/container v3 doesn't have Scoped, using Singleton as fallback
func (a *ContainerAdapter) Scoped(resolver interface{}) error {
	// golobby/container v3 doesn't support scoped lifetimes
	// Using Singleton as a fallback
	return a.container.Singleton(resolver)
}
