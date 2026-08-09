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
	default:
		return fmt.Sprintf(`{"action":"%s","result":"SUCCESS"}`, name), nil
	}
}
