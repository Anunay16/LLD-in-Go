package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"plugable_agentic_framework/pkg/models"
)

// LLMProvider is the Strategy interface that all pluggable AI model providers must implement.
type LLMProvider interface {
	// Name returns the provider identifier (e.g., "openai", "gemini", "anthropic")
	Name() string
	// GenerateResponse sends a standardized prompt/messages to the LLM and returns the completion or tool calls
	GenerateResponse(ctx context.Context, req models.LLMRequest) (*models.LLMResponse, error)
}

// ProviderRegistry is the interface abstraction for registering and retrieving LLM providers.
// High-level consumers (like Agent) depend on this interface rather than concrete registry implementations.
type ProviderRegistry interface {
	Register(provider LLMProvider)
	Get(name string) (LLMProvider, error)
	ListProviders() []string
}

// DefaultLLMRegistry is a thread-safe concrete implementation of ProviderRegistry.
type DefaultLLMRegistry struct {
	mu        sync.RWMutex
	providers map[string]LLMProvider
}

// NewLLMRegistry initializes a provider registry satisfying ProviderRegistry interface
func NewLLMRegistry() ProviderRegistry {
	return &DefaultLLMRegistry{
		providers: make(map[string]LLMProvider),
	}
}

// Register adds a provider to the registry
func (r *DefaultLLMRegistry) Register(provider LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToLower(provider.Name())] = provider
}

// Get retrieves a provider by name
func (r *DefaultLLMRegistry) Get(name string) (LLMProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not registered", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (r *DefaultLLMRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]string, 0, len(r.providers))
	for name := range r.providers {
		list = append(list, name)
	}
	return list
}

// --- Concrete Provider Implementations (Adapters) ---

// OpenAIProvider adapts OpenAI API to the LLMProvider interface
type OpenAIProvider struct {
	APIKey string
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{APIKey: apiKey}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req models.LLMRequest) (*models.LLMResponse, error) {
	lastMsg := ""
	if len(req.Messages) > 0 {
		lastMsg = strings.ToLower(req.Messages[len(req.Messages)-1].Content)
	}

	for _, m := range req.Messages {
		if m.Role == models.RoleTool {
			return &models.LLMResponse{
				Content: fmt.Sprintf("[OpenAI %s] Based on market analysis:\n%s", req.Model, m.Content),
				Usage:   models.Usage{PromptTokens: 150, CompletionTokens: 80, TotalTokens: 230},
			}, nil
		}
	}

	if strings.Contains(lastMsg, "psu") || strings.Contains(lastMsg, "performing") || strings.Contains(lastMsg, "stock") {
		var toolCall models.ToolCall
		if strings.Contains(lastMsg, "psu") {
			toolCall = models.ToolCall{
				ID:   "call_openai_psu_123",
				Name: "get_psu_stocks_analysis",
				Arguments: map[string]interface{}{
					"sector": "all",
				},
			}
		} else {
			toolCall = models.ToolCall{
				ID:   "call_openai_perf_456",
				Name: "get_top_performing_stocks",
				Arguments: map[string]interface{}{
					"limit": 5,
				},
			}
		}
		return &models.LLMResponse{
			ToolCalls: []models.ToolCall{toolCall},
			Usage:     models.Usage{PromptTokens: 100, CompletionTokens: 30, TotalTokens: 130},
		}, nil
	}

	return &models.LLMResponse{
		Content: fmt.Sprintf("[OpenAI %s] I am your stock market assistant. How can I help you analyze equities today?", req.Model),
		Usage:   models.Usage{PromptTokens: 50, CompletionTokens: 20, TotalTokens: 70},
	}, nil
}

// GeminiProvider adapts Google Gemini API to the LLMProvider interface
type GeminiProvider struct {
	APIKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{APIKey: apiKey}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) GenerateResponse(ctx context.Context, req models.LLMRequest) (*models.LLMResponse, error) {
	lastMsg := ""
	if len(req.Messages) > 0 {
		lastMsg = strings.ToLower(req.Messages[len(req.Messages)-1].Content)
	}

	for _, m := range req.Messages {
		if m.Role == models.RoleTool {
			return &models.LLMResponse{
				Content: fmt.Sprintf("[Gemini %s Insights] Synthesis of data:\n%s\nRecommendation: Maintain balanced asset allocation.", req.Model, m.Content),
				Usage:   models.Usage{PromptTokens: 140, CompletionTokens: 75, TotalTokens: 215},
			}, nil
		}
	}

	if strings.Contains(lastMsg, "psu") || strings.Contains(lastMsg, "performing") || strings.Contains(lastMsg, "stock") {
		var toolCall models.ToolCall
		if strings.Contains(lastMsg, "psu") {
			toolCall = models.ToolCall{
				ID:   "gemini_call_psu_789",
				Name: "get_psu_stocks_analysis",
				Arguments: map[string]interface{}{
					"sector": "energy_and_banking",
				},
			}
		} else {
			toolCall = models.ToolCall{
				ID:   "gemini_call_perf_999",
				Name: "get_top_performing_stocks",
				Arguments: map[string]interface{}{
					"limit": 3,
				},
			}
		}
		return &models.LLMResponse{
			ToolCalls: []models.ToolCall{toolCall},
			Usage:     models.Usage{PromptTokens: 95, CompletionTokens: 25, TotalTokens: 120},
		}, nil
	}

	return &models.LLMResponse{
		Content: fmt.Sprintf("[Gemini %s] Ready to analyze stock trends and PSU benchmarks.", req.Model),
		Usage:   models.Usage{PromptTokens: 40, CompletionTokens: 15, TotalTokens: 55},
	}, nil
}

// AnthropicProvider adapts Anthropic Claude API to the LLMProvider interface
type AnthropicProvider struct {
	APIKey string
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{APIKey: apiKey}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) GenerateResponse(ctx context.Context, req models.LLMRequest) (*models.LLMResponse, error) {
	lastMsg := ""
	if len(req.Messages) > 0 {
		lastMsg = strings.ToLower(req.Messages[len(req.Messages)-1].Content)
	}

	for _, m := range req.Messages {
		if m.Role == models.RoleTool {
			return &models.LLMResponse{
				Content: fmt.Sprintf("[Claude %s Financial Analysis]\nHere is the detailed evaluation:\n%s", req.Model, m.Content),
				Usage:   models.Usage{PromptTokens: 160, CompletionTokens: 90, TotalTokens: 250},
			}, nil
		}
	}

	if strings.Contains(lastMsg, "psu") || strings.Contains(lastMsg, "performing") || strings.Contains(lastMsg, "stock") {
		var toolCall models.ToolCall
		if strings.Contains(lastMsg, "psu") {
			toolCall = models.ToolCall{
				ID:   "anthropic_call_psu_101",
				Name: "get_psu_stocks_analysis",
				Arguments: map[string]interface{}{
					"sector": "defence_and_power",
				},
			}
		} else {
			toolCall = models.ToolCall{
				ID:   "anthropic_call_perf_202",
				Name: "get_top_performing_stocks",
				Arguments: map[string]interface{}{
					"limit": 5,
				},
			}
		}
		return &models.LLMResponse{
			ToolCalls: []models.ToolCall{toolCall},
			Usage:     models.Usage{PromptTokens: 105, CompletionTokens: 35, TotalTokens: 140},
		}, nil
	}

	return &models.LLMResponse{
		Content: fmt.Sprintf("[Claude %s] Hello! I can assist with fundamental and technical stock evaluation.", req.Model),
		Usage:   models.Usage{PromptTokens: 45, CompletionTokens: 18, TotalTokens: 63},
	}, nil
}
