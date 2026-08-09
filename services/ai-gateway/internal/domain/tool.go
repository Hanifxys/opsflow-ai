package domain

type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"` // If true, requires human approval before execution
}

type ToolResult struct {
	ToolCallID       string `json:"tool_call_id"`
	ToolName         string `json:"tool_name"`
	Output           string `json:"output"`
	RequiresApproval bool   `json:"requires_approval"`
	ApprovalID       string `json:"approval_id,omitempty"`
}
