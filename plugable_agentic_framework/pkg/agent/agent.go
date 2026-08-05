package agent

import (
	"context"
	"fmt"
	"time"

	"plugable_agentic_framework/pkg/llm"
	"plugable_agentic_framework/pkg/models"
	"plugable_agentic_framework/pkg/tools"
)

// Agent is the core interface for executing agentic queries
type Agent interface {
	ExecuteQuery(ctx context.Context, req models.AgentRequest) (*models.AgentResponse, error)
}

// PluggableStockAgent orchestrates LLM providers and financial tools.
// Following Dependency Inversion Principle (DIP), it depends on interfaces
// (llm.ProviderRegistry and tools.ToolRegistry) rather than concrete implementations.
type PluggableStockAgent struct {
	providerRegistry llm.ProviderRegistry
	toolRegistry     tools.ToolRegistry
	systemPrompt     string
}

// NewPluggableStockAgent initializes the agent with injected interface abstractions
func NewPluggableStockAgent(providerReg llm.ProviderRegistry, toolReg tools.ToolRegistry) *PluggableStockAgent {
	return &PluggableStockAgent{
		providerRegistry: providerReg,
		toolRegistry:     toolReg,
		systemPrompt:     "You are an expert AI Stock Market Assistant. Help users analyze equity performance, suggest high-performing stocks, and evaluate Indian PSU (Public Sector Undertaking) opportunities using available market tools.",
	}
}

// ExecuteQuery processes a query using the requested LLM provider and available tools
func (a *PluggableStockAgent) ExecuteQuery(ctx context.Context, req models.AgentRequest) (*models.AgentResponse, error) {
	startTime := time.Now()

	// 1. Fetch requested pluggable LLM provider abstraction via ProviderRegistry interface
	provider, err := a.providerRegistry.Get(req.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("failed to load LLM provider: %w", err)
	}

	// 2. Set default iterations limit if not specified
	maxIterations := req.MaxToolIterations
	if maxIterations <= 0 {
		maxIterations = 5
	}

	// 3. Build initial conversation state
	conversation := []models.Message{
		{Role: models.RoleSystem, Content: a.systemPrompt},
		{Role: models.RoleUser, Content: req.Query},
	}

	// 4. Retrieve definitions for all registered tools via ToolRegistry interface
	toolDefs := a.toolRegistry.Definitions()

	executedTools := make([]string, 0)

	// 5. Execution Loop (ReAct / Tool Call Orchestration)
	for i := 0; i < maxIterations; i++ {
		llmReq := models.LLMRequest{
			Model:       req.ModelName,
			Messages:    conversation,
			Tools:       toolDefs,
			Temperature: 0.2,
		}

		resp, err := provider.GenerateResponse(ctx, llmReq)
		if err != nil {
			return nil, fmt.Errorf("LLM provider execution error: %w", err)
		}

		// Case A: Model wants to execute one or more tools
		if len(resp.ToolCalls) > 0 {
			for _, call := range resp.ToolCalls {
				executedTools = append(executedTools, call.Name)

				tool, exists := a.toolRegistry.Get(call.Name)
				var toolOutput string
				if !exists {
					toolOutput = fmt.Sprintf("Error: tool '%s' not registered", call.Name)
				} else {
					out, err := tool.Execute(ctx, call.Arguments)
					if err != nil {
						toolOutput = fmt.Sprintf("Error executing tool '%s': %v", call.Name, err)
					} else {
						toolOutput = out
					}
				}

				// Append tool execution turn to conversation history
				conversation = append(conversation, models.Message{
					Role:       models.RoleTool,
					Content:    toolOutput,
					ToolCallID: call.ID,
				})
			}
			continue
		}

		// Case B: Model completed answer synthesis
		return &models.AgentResponse{
			Query:         req.Query,
			Answer:        resp.Content,
			ProviderUsed:  provider.Name(),
			ModelUsed:     req.ModelName,
			ToolsExecuted: executedTools,
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	return nil, fmt.Errorf("agent exceeded max tool iterations (%d)", maxIterations)
}
