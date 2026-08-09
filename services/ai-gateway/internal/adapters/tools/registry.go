package tools

import (
	"context"
	"fmt"

	"github.com/opsflow/ai-gateway/internal/domain"
)

type ToolRegistry struct {
	tools map[string]domain.ToolDefinition
}

func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		tools: make(map[string]domain.ToolDefinition),
	}
	r.registerDefaultTools()
	return r
}

func (r *ToolRegistry) registerDefaultTools() {
	// Standard operational tools
	r.tools["get_service_status"] = domain.ToolDefinition{
		Name:        "get_service_status",
		Description: "Fetches current health status and environment details of a registered service",
		Sensitive:   false,
	}
	r.tools["list_incidents"] = domain.ToolDefinition{
		Name:        "list_incidents",
		Description: "Lists active and recent incidents across system microservices",
		Sensitive:   false,
	}
	r.tools["restart_service"] = domain.ToolDefinition{
		Name:        "restart_service",
		Description: "Restarts a target operational microservice deployment",
		Sensitive:   true, // Requires human approval!
	}
	r.tools["scale_deployment"] = domain.ToolDefinition{
		Name:        "scale_deployment",
		Description: "Scales total replica count of a target service deployment",
		Sensitive:   true, // Requires human approval!
	}

	// L3 Support & Infrastructure Engineering Specialized Tools
	r.tools["analyze_stacktrace"] = domain.ToolDefinition{
		Name:        "analyze_stacktrace",
		Description: "Deeply analyzes raw error log stacktraces to determine root cause and recommended remediation",
		Sensitive:   false,
	}
	r.tools["run_database_diagnostics"] = domain.ToolDefinition{
		Name:        "run_database_diagnostics",
		Description: "Executes deep PostgreSQL database pool, active queries, and lock contention diagnostics",
		Sensitive:   false,
	}
	r.tools["generate_rca_report"] = domain.ToolDefinition{
		Name:        "generate_rca_report",
		Description: "Generates a comprehensive Root Cause Analysis (RCA) report for active or past incidents",
		Sensitive:   false,
	}
	r.tools["flush_redis_cache"] = domain.ToolDefinition{
		Name:        "flush_redis_cache",
		Description: "Flushes target Redis cache keys space for a microservice",
		Sensitive:   true, // Requires human approval!
	}
}

func (r *ToolRegistry) Definitions() []domain.ToolDefinition {
	var list []domain.ToolDefinition
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *ToolRegistry) IsSensitive(toolName string) bool {
	if t, exists := r.tools[toolName]; exists {
		return t.Sensitive
	}
	return true // Default to sensitive for unknown tools!
}

func (r *ToolRegistry) ExecuteReadOnlyTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	if r.IsSensitive(name) {
		return "", domain.ErrToolExecutionDenied
	}

	switch name {
	case "get_service_status":
		svc := fmt.Sprintf("%v", args["service_name"])
		return fmt.Sprintf(`{"service":"%s","status":"HEALTHY","latency_ms":12,"replicas":3}`, svc), nil
	case "list_incidents":
		return `[{"id":"inc-101","title":"Payment Gateway Timeout","severity":"HIGH","status":"INVESTIGATING"}]`, nil
	case "analyze_stacktrace":
		return `{"root_cause":"pq: connection pool max limit (100) exceeded in postgres_driver.go:84","affected_component":"payment-service","severity":"CRITICAL","recommendation":"Scale connection pool max_conns to 200 or add Redis read cache"}`, nil
	case "run_database_diagnostics":
		return `{"database":"opsflow","active_connections":98,"max_connections":100,"idle_connections":2,"slowest_query":"SELECT * FROM incidents WHERE status = 'OPEN' FOR UPDATE","pool_status":"WARNING_EXHAUSTED"}`, nil
	case "generate_rca_report":
		return `# Root Cause Analysis (RCA) — INC-2026-001\n\n## Incident Summary\nPayment service experienced connection pool exhaustion resulting in HTTP 504 errors.\n\n## Root Cause\nUnindexed SELECT FOR UPDATE query in incident repository combined with sudden traffic spike.\n\n## Remediation\n1. Added composite index on outbox_events(status, created_at).\n2. Configured Redis read cache for service registry.\n3. Increased pgx pool max_conns.`, nil
	default:
		return fmt.Sprintf(`{"tool":"%s","result":"executed"}`, name), nil
	}
}

func (r *ToolRegistry) ExecuteApprovedMutation(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	switch name {
	case "restart_service":
		svc := fmt.Sprintf("%v", args["service_name"])
		return fmt.Sprintf(`{"action":"restart_service","service":"%s","result":"SUCCESS","restarted_at":"NOW"}`, svc), nil
	case "scale_deployment":
		svc := fmt.Sprintf("%v", args["service_name"])
		return fmt.Sprintf(`{"action":"scale_deployment","service":"%s","result":"SUCCESS"}`, svc), nil
	case "flush_redis_cache":
		svc := fmt.Sprintf("%v", args["service_name"])
		return fmt.Sprintf(`{"action":"flush_redis_cache","service":"%s","flushed_keys":42,"result":"SUCCESS"}`, svc), nil
	default:
		return fmt.Sprintf(`{"action":"%s","result":"SUCCESS"}`, name), nil
	}
}
