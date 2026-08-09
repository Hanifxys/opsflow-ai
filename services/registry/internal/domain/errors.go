package domain

import "errors"

var (
	ErrServiceNotFound     = errors.New("service not found")
	ErrServiceNameExists   = errors.New("service name already exists")
	ErrEnvironmentExists   = errors.New("environment already exists for service")
	ErrEnvironmentNotFound = errors.New("service environment not found")
	ErrDependencyExists    = errors.New("dependency already exists")
	ErrSelfDependency      = errors.New("service cannot depend on itself")
	ErrHealthCheckNotFound = errors.New("health check definition not found")
)
