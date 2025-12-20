package ioc

// Container defines the interface for dependency injection container
// This abstraction allows for better testability and flexibility
type Container interface {
	// Resolve resolves a dependency by its type
	// The target must be a pointer to the type you want to resolve
	Resolve(target interface{}) error

	// Singleton registers a singleton dependency
	Singleton(resolver interface{}) error

	// Transient registers a transient dependency (new instance each time)
	Transient(resolver interface{}) error

	// Scoped registers a scoped dependency (one instance per scope)
	Scoped(resolver interface{}) error
}

// Ensure ContainerBuilder implements Container interface
var _ Container = (*ContainerBuilder)(nil)
