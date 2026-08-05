package models

import "time"

// Role defines the message sender role
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single turn in a conversation
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolDefinition defines the metadata and parameter schema for an agent tool
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema properties
}

// ToolCall represents a model's request to execute a specific tool
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult contains the output of executing a tool
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Content    string `json:"content"`
	Error      error  `json:"error,omitempty"`
}

// LLMRequest encapsulates the request sent to any pluggable LLM provider
type LLMRequest struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	Temperature float64          `json:"temperature"`
}

// LLMResponse is the standardized output from any pluggable LLM provider
type LLMResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     Usage      `json:"usage"`
}

// Usage tracks token counts
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AgentRequest is the user input to the stock agent
type AgentRequest struct {
	Query            string `json:"query"`
	ProviderName     string `json:"provider_name"`     // e.g. "openai", "gemini", "anthropic"
	ModelName        string `json:"model_name"`        // e.g. "gpt-4o", "gemini-1.5-pro", "claude-3-5-sonnet"
	MaxToolIterations int    `json:"max_tool_iterations"`
}

// AgentResponse is the output delivered back to the user
type AgentResponse struct {
	Query         string        `json:"query"`
	Answer        string        `json:"answer"`
	ProviderUsed  string        `json:"provider_used"`
	ModelUsed     string        `json:"model_used"`
	ToolsExecuted []string      `json:"tools_executed"`
	ExecutionTime time.Duration `json:"execution_time"`
}
