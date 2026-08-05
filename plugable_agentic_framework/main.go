package main

import (
	"context"
	"fmt"
	"log"

	"plugable_agentic_framework/pkg/agent"
	"plugable_agentic_framework/pkg/llm"
	"plugable_agentic_framework/pkg/models"
	"plugable_agentic_framework/pkg/tools"
)

func main() {
	fmt.Println("==========================================================================")
	fmt.Println("         PLUGGABLE AI AGENT FRAMEWORK - STOCK MARKET DEMO                ")
	fmt.Println("==========================================================================")
	fmt.Println()

	// 1. Initialize LLM Registry & Register Multiple Pluggable Model Providers
	llmRegistry := llm.NewLLMRegistry()
	llmRegistry.Register(llm.NewOpenAIProvider("sk-fake-openai-key"))
	llmRegistry.Register(llm.NewGeminiProvider("fake-gemini-key"))
	llmRegistry.Register(llm.NewAnthropicProvider("sk-ant-fake-key"))

	fmt.Printf("Registered LLM Providers: %v\n\n", llmRegistry.ListProviders())

	// 2. Initialize Tool Registry & Register Pluggable Stock Tools
	toolRegistry := tools.NewToolRegistry()
	toolRegistry.Register(tools.NewPSUSuggestionTool())
	toolRegistry.Register(tools.NewTopPerformingStocksTool())

	// 3. Instantiate the Pluggable Agent
	stockAgent := agent.NewPluggableStockAgent(llmRegistry, toolRegistry)

	ctx := context.Background()

	// --- Test Scenario 1: PSU Stock Suggestion using OpenAI ---
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Println("Scenario 1: Querying PSU stocks using OpenAI (gpt-4o)")
	fmt.Println("--------------------------------------------------------------------------")
	req1 := models.AgentRequest{
		Query:        "Can you suggest me some good PSU stocks to invest in for long term?",
		ProviderName: "openai",
		ModelName:    "gpt-4o",
	}

	resp1, err := stockAgent.ExecuteQuery(ctx, req1)
	if err != nil {
		log.Fatalf("Error executing query 1: %v", err)
	}
	printResponse(resp1)

	// --- Test Scenario 2: Same PSU Stock Query using Google Gemini ---
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Println("Scenario 2: SWAPPING PROVIDER -> Querying PSU stocks using Gemini (gemini-1.5-pro)")
	fmt.Println("--------------------------------------------------------------------------")
	req2 := models.AgentRequest{
		Query:        "Suggest some PSU stocks and explain why they are performing good.",
		ProviderName: "gemini",
		ModelName:    "gemini-1.5-pro",
	}

	resp2, err := stockAgent.ExecuteQuery(ctx, req2)
	if err != nil {
		log.Fatalf("Error executing query 2: %v", err)
	}
	printResponse(resp2)

	// --- Test Scenario 3: General Stock Performance Query using Anthropic ---
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Println("Scenario 3: SWAPPING PROVIDER -> Querying Top Performing Stocks using Anthropic (claude-3-5-sonnet)")
	fmt.Println("--------------------------------------------------------------------------")
	req3 := models.AgentRequest{
		Query:        "Which stocks are performing good in the market right now?",
		ProviderName: "anthropic",
		ModelName:    "claude-3-5-sonnet",
	}

	resp3, err := stockAgent.ExecuteQuery(ctx, req3)
	if err != nil {
		log.Fatalf("Error executing query 3: %v", err)
	}
	printResponse(resp3)
}

func printResponse(resp *models.AgentResponse) {
	fmt.Printf("Query:          %s\n", resp.Query)
	fmt.Printf("Provider Used:  %s\n", resp.ProviderUsed)
	fmt.Printf("Model Used:     %s\n", resp.ModelUsed)
	fmt.Printf("Tools Executed: %v\n", resp.ToolsExecuted)
	fmt.Printf("Latency:        %v\n", resp.ExecutionTime)
	fmt.Println("Answer:")
	fmt.Println(resp.Answer)
	fmt.Println()
}
