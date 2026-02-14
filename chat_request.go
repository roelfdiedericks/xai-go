package xai

import (
	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// ReasoningEffort controls how much reasoning effort the model should use.
type ReasoningEffort int

const (
	// ReasoningEffortLow uses minimal reasoning.
	ReasoningEffortLow ReasoningEffort = iota + 1
	// ReasoningEffortMedium uses moderate reasoning (default).
	ReasoningEffortMedium
	// ReasoningEffortHigh uses maximum reasoning effort.
	ReasoningEffortHigh
)

func (r ReasoningEffort) toProto() v1.ReasoningEffort {
	switch r {
	case ReasoningEffortLow:
		return v1.ReasoningEffort_EFFORT_LOW
	case ReasoningEffortMedium:
		return v1.ReasoningEffort_EFFORT_MEDIUM
	case ReasoningEffortHigh:
		return v1.ReasoningEffort_EFFORT_HIGH
	default:
		return v1.ReasoningEffort_INVALID_EFFORT
	}
}

// ResponseFormat controls the format of the model's response.
type ResponseFormat int

const (
	// ResponseFormatText returns plain text (default).
	ResponseFormatText ResponseFormat = iota
	// ResponseFormatJSON returns JSON output.
	ResponseFormatJSON
)

// ChatRequest builds a chat completion request.
type ChatRequest struct {
	messages            []*v1.Message
	model               string
	user                string
	maxTokens           *int32
	seed                *int32
	stop                []string
	temperature         *float32
	topP                *float32
	logprobs            bool
	topLogprobs         *int32
	tools               []Tool
	toolChoice          *ToolChoice
	responseFormat      *ResponseFormat
	frequencyPenalty    *float32
	presencePenalty     *float32
	reasoningEffort     *ReasoningEffort
	parallelToolCalls   *bool
	storeMessages       bool
	maxTurns            *int32
	includeOptions      []v1.IncludeOption
	previousResponseID  string
	useEncryptedContent bool
}

// NewChatRequest creates a new empty chat request builder.
func NewChatRequest() *ChatRequest {
	return &ChatRequest{}
}

// SystemMessage adds a system message to the conversation.
func (r *ChatRequest) SystemMessage(text string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role: v1.MessageRole_ROLE_SYSTEM,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
		},
	})
	return r
}

// UserMessage adds a user message to the conversation.
func (r *ChatRequest) UserMessage(text string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role: v1.MessageRole_ROLE_USER,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
		},
	})
	return r
}

// UserWithImage adds a user message with text and an image URL.
func (r *ChatRequest) UserWithImage(text, imageURL string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role: v1.MessageRole_ROLE_USER,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
			{Content: &v1.Content_ImageUrl{ImageUrl: &v1.ImageUrlContent{ImageUrl: imageURL}}},
		},
	})
	return r
}

// AssistantMessage adds an assistant message to the conversation.
func (r *ChatRequest) AssistantMessage(text string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role: v1.MessageRole_ROLE_ASSISTANT,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
		},
	})
	return r
}

// ToolResult adds a tool result message to the conversation.
func (r *ChatRequest) ToolResult(toolCallID, result string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role:       v1.MessageRole_ROLE_TOOL,
		ToolCallId: &toolCallID,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: result}},
		},
	})
	return r
}

// DeveloperMessage adds a developer instruction message.
func (r *ChatRequest) DeveloperMessage(text string) *ChatRequest {
	r.messages = append(r.messages, &v1.Message{
		Role: v1.MessageRole_ROLE_DEVELOPER,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
		},
	})
	return r
}

// WithModel sets the model to use.
func (r *ChatRequest) WithModel(model string) *ChatRequest {
	r.model = model
	return r
}

// WithUser sets an opaque user identifier for logging.
func (r *ChatRequest) WithUser(user string) *ChatRequest {
	r.user = user
	return r
}

// WithMaxTokens sets the maximum number of tokens to generate.
func (r *ChatRequest) WithMaxTokens(n int32) *ChatRequest {
	r.maxTokens = &n
	return r
}

// WithSeed sets a random seed for deterministic sampling.
func (r *ChatRequest) WithSeed(seed int32) *ChatRequest {
	r.seed = &seed
	return r
}

// WithStop sets stop sequences that will terminate generation.
func (r *ChatRequest) WithStop(sequences ...string) *ChatRequest {
	r.stop = sequences
	return r
}

// WithTemperature sets the sampling temperature (0-2).
// Lower values make output more deterministic.
func (r *ChatRequest) WithTemperature(t float32) *ChatRequest {
	r.temperature = &t
	return r
}

// WithTopP sets the nucleus sampling parameter (0-1).
func (r *ChatRequest) WithTopP(p float32) *ChatRequest {
	r.topP = &p
	return r
}

// WithLogprobs enables log probability output.
func (r *ChatRequest) WithLogprobs(topLogprobs int32) *ChatRequest {
	r.logprobs = true
	r.topLogprobs = &topLogprobs
	return r
}

// AddTool adds a tool to the request.
func (r *ChatRequest) AddTool(tool Tool) *ChatRequest {
	r.tools = append(r.tools, tool)
	return r
}

// AddTools adds multiple tools to the request.
func (r *ChatRequest) AddTools(tools ...Tool) *ChatRequest {
	r.tools = append(r.tools, tools...)
	return r
}

// WithToolChoice sets the tool choice mode.
func (r *ChatRequest) WithToolChoice(choice ToolChoice) *ChatRequest {
	r.toolChoice = &choice
	return r
}

// WithResponseFormat sets the response format.
func (r *ChatRequest) WithResponseFormat(format ResponseFormat) *ChatRequest {
	r.responseFormat = &format
	return r
}

// WithFrequencyPenalty sets the frequency penalty (-2 to 2).
func (r *ChatRequest) WithFrequencyPenalty(p float32) *ChatRequest {
	r.frequencyPenalty = &p
	return r
}

// WithPresencePenalty sets the presence penalty (-2 to 2).
func (r *ChatRequest) WithPresencePenalty(p float32) *ChatRequest {
	r.presencePenalty = &p
	return r
}

// WithReasoningEffort sets the reasoning effort for reasoning models.
func (r *ChatRequest) WithReasoningEffort(effort ReasoningEffort) *ChatRequest {
	r.reasoningEffort = &effort
	return r
}

// WithParallelToolCalls controls whether tools can be called in parallel.
func (r *ChatRequest) WithParallelToolCalls(enabled bool) *ChatRequest {
	r.parallelToolCalls = &enabled
	return r
}

// WithStoreMessages enables storing messages on xAI's servers for later retrieval.
// When enabled, you can use the response ID with WithPreviousResponseId to continue
// conversations without resending the full history. Messages are stored for 30 days.
func (r *ChatRequest) WithStoreMessages(store bool) *ChatRequest {
	r.storeMessages = store
	return r
}

// WithPreviousResponseId continues a conversation from a previous response.
// When set, only new messages need to be added - the server will chain them
// with the stored conversation history. Requires that the previous request
// was made with WithStoreMessages(true).
func (r *ChatRequest) WithPreviousResponseId(id string) *ChatRequest {
	r.previousResponseID = id
	return r
}

// WithEncryptedContent enables encrypted thinking content for reasoning models.
// This allows reasoning traces to be preserved and rehydrated across conversation
// turns when using WithPreviousResponseId.
func (r *ChatRequest) WithEncryptedContent(enabled bool) *ChatRequest {
	r.useEncryptedContent = enabled
	return r
}

// WithMaxTurns sets the maximum number of agentic tool calling turns.
func (r *ChatRequest) WithMaxTurns(n int32) *ChatRequest {
	r.maxTurns = &n
	return r
}

// IncludeWebSearchOutput includes encrypted web search tool output.
func (r *ChatRequest) IncludeWebSearchOutput() *ChatRequest {
	r.includeOptions = append(r.includeOptions, v1.IncludeOption_INCLUDE_OPTION_WEB_SEARCH_CALL_OUTPUT)
	return r
}

// IncludeXSearchOutput includes encrypted X search tool output.
func (r *ChatRequest) IncludeXSearchOutput() *ChatRequest {
	r.includeOptions = append(r.includeOptions, v1.IncludeOption_INCLUDE_OPTION_X_SEARCH_CALL_OUTPUT)
	return r
}

// IncludeCodeExecutionOutput includes code execution tool output.
func (r *ChatRequest) IncludeCodeExecutionOutput() *ChatRequest {
	r.includeOptions = append(r.includeOptions, v1.IncludeOption_INCLUDE_OPTION_CODE_EXECUTION_CALL_OUTPUT)
	return r
}

// IncludeInlineCitations includes inline citations in the response.
func (r *ChatRequest) IncludeInlineCitations() *ChatRequest {
	r.includeOptions = append(r.includeOptions, v1.IncludeOption_INCLUDE_OPTION_INLINE_CITATIONS)
	return r
}

// IncludeVerboseStreaming streams all chunks including those without user content.
func (r *ChatRequest) IncludeVerboseStreaming() *ChatRequest {
	r.includeOptions = append(r.includeOptions, v1.IncludeOption_INCLUDE_OPTION_VERBOSE_STREAMING)
	return r
}

// Build converts the request to a proto message.
// If model is not set, it uses the provided default model.
func (r *ChatRequest) Build(defaultModel string) *v1.GetCompletionsRequest {
	req := &v1.GetCompletionsRequest{
		Messages:            r.messages,
		Model:               r.model,
		User:                r.user,
		Stop:                r.stop,
		Logprobs:            r.logprobs,
		StoreMessages:       r.storeMessages,
		Include:             r.includeOptions,
		UseEncryptedContent: r.useEncryptedContent,
	}

	// Previous response ID for conversation continuation
	if r.previousResponseID != "" {
		req.PreviousResponseId = &r.previousResponseID
	}

	// Use default model if not specified
	if req.Model == "" {
		req.Model = defaultModel
	}

	// Optional fields
	if r.maxTokens != nil {
		req.MaxTokens = r.maxTokens
	}
	if r.seed != nil {
		req.Seed = r.seed
	}
	if r.temperature != nil {
		req.Temperature = r.temperature
	}
	if r.topP != nil {
		req.TopP = r.topP
	}
	if r.topLogprobs != nil {
		req.TopLogprobs = r.topLogprobs
	}
	if r.frequencyPenalty != nil {
		req.FrequencyPenalty = r.frequencyPenalty
	}
	if r.presencePenalty != nil {
		req.PresencePenalty = r.presencePenalty
	}
	if r.reasoningEffort != nil {
		effort := r.reasoningEffort.toProto()
		req.ReasoningEffort = &effort
	}
	if r.parallelToolCalls != nil {
		req.ParallelToolCalls = r.parallelToolCalls
	}
	if r.maxTurns != nil {
		req.MaxTurns = r.maxTurns
	}

	// Tools
	for _, tool := range r.tools {
		req.Tools = append(req.Tools, tool.toProto())
	}

	// Tool choice
	if r.toolChoice != nil {
		req.ToolChoice = r.toolChoice.toProto()
	}

	// Response format
	if r.responseFormat != nil {
		switch *r.responseFormat {
		case ResponseFormatJSON:
			req.ResponseFormat = &v1.ResponseFormat{
				FormatType: v1.FormatType_FORMAT_TYPE_JSON_OBJECT,
			}
		default:
			req.ResponseFormat = &v1.ResponseFormat{
				FormatType: v1.FormatType_FORMAT_TYPE_TEXT,
			}
		}
	}

	return req
}

// Messages returns the current messages in the request.
func (r *ChatRequest) Messages() []*v1.Message {
	return r.messages
}

// Tools returns the current tools in the request.
func (r *ChatRequest) Tools() []Tool {
	return r.tools
}
