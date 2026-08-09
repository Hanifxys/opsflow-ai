package domain

import "errors"

var (
	ErrIncidentNotFound  = errors.New("incident not found")
	ErrInvalidTransition = errors.New("invalid status transition for incident")
	ErrInvalidSeverity   = errors.New("invalid incident severity")
	ErrInvalidStatus     = errors.New("invalid incident status")
)
