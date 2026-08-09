module github.com/opsflow/tests/integration

go 1.22.0

require (
	github.com/google/uuid v1.6.0
	github.com/opsflow/ai-gateway v0.0.0
	github.com/opsflow/auth-service v0.0.0
	github.com/opsflow/common v0.0.0
	github.com/opsflow/incident-service v0.0.0
	github.com/opsflow/notification-service v0.0.0
	github.com/opsflow/registry-service v0.0.0
)

require github.com/golang-jwt/jwt/v5 v5.2.1 // indirect

replace (
	github.com/opsflow/ai-gateway => ../../services/ai-gateway
	github.com/opsflow/auth-service => ../../services/auth
	github.com/opsflow/common => ../../pkg/common
	github.com/opsflow/incident-service => ../../services/incident
	github.com/opsflow/notification-service => ../../services/notification
	github.com/opsflow/registry-service => ../../services/registry
)
