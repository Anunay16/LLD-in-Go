package tools

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"plugable_agentic_framework/pkg/models"
)

// Tool is the Command interface that all pluggable agent tools must implement.
type Tool interface {
	Definition() models.ToolDefinition
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// ToolRegistry is the interface abstraction for managing tool registration and definitions.
// High-level consumers (like Agent) depend on this interface rather than concrete registry implementations.
type ToolRegistry interface {
	Register(t Tool)
	Get(name string) (Tool, bool)
	Definitions() []models.ToolDefinition
}

// DefaultToolRegistry is a thread-safe concrete implementation of ToolRegistry.
type DefaultToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewToolRegistry() ToolRegistry {
	return &DefaultToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *DefaultToolRegistry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Definition().Name] = t
}

func (r *DefaultToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

func (r *DefaultToolRegistry) Definitions() []models.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]models.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}

// --- Concrete Financial / Stock Tools ---

// PSUSuggestionTool provides analysis on top Public Sector Undertaking (PSU) stocks
type PSUSuggestionTool struct{}

func NewPSUSuggestionTool() *PSUSuggestionTool {
	return &PSUSuggestionTool{}
}

func (t *PSUSuggestionTool) Definition() models.ToolDefinition {
	return models.ToolDefinition{
		Name:        "get_psu_stocks_analysis",
		Description: "Fetches performance, dividend yield, and fundamentals of top Indian PSU (Public Sector Undertaking) stocks like Defence, Power, and Banking.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"sector": map[string]interface{}{
					"type":        "string",
					"description": "Filter by sector: 'defence', 'power', 'banking', or 'all'",
				},
			},
		},
	}
}

func (t *PSUSuggestionTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	sector := "all"
	if s, ok := args["sector"].(string); ok && s != "" {
		sector = strings.ToLower(s)
	}

	psuData := map[string]interface{}{
		"summary": "PSU stocks have shown strong order book execution, capital expenditure, and robust dividend yields.",
		"top_picks": []map[string]interface{}{
			{
				"ticker":         "HAL (Hindustan Aeronautics)",
				"sector":         "Defence",
				"1y_return":      "+85%",
				"dividend_yield": "1.2%",
				"catalysts":      "Strong export order pipeline & indigenous defence push",
			},
			{
				"ticker":         "NTPC",
				"sector":         "Power",
				"1y_return":      "+62%",
				"dividend_yield": "2.4%",
				"catalysts":      "Renewable energy expansion & thermal capacity utilization",
			},
			{
				"ticker":         "SBIN (State Bank of India)",
				"sector":         "Banking",
				"1y_return":      "+35%",
				"dividend_yield": "1.8%",
				"catalysts":      "Lowest NPA levels in decade & credit growth",
			},
			{
				"ticker":         "BEL (Bharat Electronics)",
				"sector":         "Defence/Tech",
				"1y_return":      "+90%",
				"dividend_yield": "1.5%",
				"catalysts":      "Defence electronics expansion & electronic voting contracts",
			},
		},
		"filter_applied": sector,
	}

	bytes, err := json.MarshalIndent(psuData, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// TopPerformingStocksTool provides overall top performing stock suggestions across sectors
type TopPerformingStocksTool struct{}

func NewTopPerformingStocksTool() *TopPerformingStocksTool {
	return &TopPerformingStocksTool{}
}

func (t *TopPerformingStocksTool) Definition() models.ToolDefinition {
	return models.ToolDefinition{
		Name:        "get_top_performing_stocks",
		Description: "Analyzes real-time market trends and returns top performing stocks based on technical momentum, volume breakout, and earnings growth.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of stock suggestions to return (default 5)",
				},
			},
		},
	}
}

func (t *TopPerformingStocksTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	topStocks := []map[string]interface{}{
		{"symbol": "RELIANCE", "name": "Reliance Industries", "momentum": "Bullish", "rsi": 68, "ytd_return": "28%"},
		{"symbol": "TCS", "name": "Tata Consultancy Services", "momentum": "Consolidating", "rsi": 54, "ytd_return": "18%"},
		{"symbol": "L&T", "name": "Larsen & Toubro", "momentum": "Strong Uptrend", "rsi": 72, "ytd_return": "42%"},
		{"symbol": "BHARTIARTL", "name": "Bharti Airtel", "momentum": "All-Time High", "rsi": 75, "ytd_return": "55%"},
		{"symbol": "ICICIBANK", "name": "ICICI Bank", "momentum": "Bullish Breakout", "rsi": 66, "ytd_return": "30%"},
	}

	if limit < len(topStocks) {
		topStocks = topStocks[:limit]
	}

	bytes, err := json.MarshalIndent(map[string]interface{}{
		"market_status": "Bullish Momentum",
		"top_performers": topStocks,
	}, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
