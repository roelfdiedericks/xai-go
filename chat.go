package xai

import (
	"context"
	"io"
	"time"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// FinishReason indicates why the model stopped generating.
type FinishReason string

const (
	// FinishReasonStop indicates the model hit a stop sequence or end of message.
	FinishReasonStop FinishReason = "stop"
	// FinishReasonLength indicates the model hit the max token limit.
	FinishReasonLength FinishReason = "length"
	// FinishReasonToolCalls indicates the model wants to call tools.
	FinishReasonToolCalls FinishReason = "tool_calls"
	// FinishReasonContentFilter indicates the content was filtered.
	FinishReasonContentFilter FinishReason = "content_filter"
)

func finishReasonFromProto(r v1.FinishReason) FinishReason {
	switch r {
	case v1.FinishReason_REASON_STOP:
		return FinishReasonStop
	case v1.FinishReason_REASON_MAX_LEN, v1.FinishReason_REASON_MAX_CONTEXT:
		return FinishReasonLength
	case v1.FinishReason_REASON_TOOL_CALLS:
		return FinishReasonToolCalls
	default:
		return ""
	}
}

// Usage contains token usage information.
type Usage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int32
	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int32
	// TotalTokens is the total number of tokens.
	TotalTokens int32
	// ReasoningTokens is the number of tokens used for reasoning.
	ReasoningTokens int32
	// CachedPromptTokens is the number of cached prompt tokens.
	CachedPromptTokens int32
	// PromptTextTokens is the number of text tokens in the prompt.
	PromptTextTokens int32
	// PromptImageTokens is the number of image tokens in the prompt.
	PromptImageTokens int32
}

func usageFromProto(u *v1.SamplingUsage) Usage {
	if u == nil {
		return Usage{}
	}
	return Usage{
		PromptTokens:       u.GetPromptTokens(),
		CompletionTokens:   u.GetCompletionTokens(),
		TotalTokens:        u.GetTotalTokens(),
		ReasoningTokens:    u.GetReasoningTokens(),
		CachedPromptTokens: u.GetCachedPromptTextTokens(),
		PromptTextTokens:   u.GetPromptTextTokens(),
		PromptImageTokens:  u.GetPromptImageTokens(),
	}
}

// ChatResponse represents a complete chat response.
type ChatResponse struct {
	// ID is the unique identifier for this response.
	ID string
	// Content is the generated text content.
	Content string
	// ReasoningContent is the reasoning trace (if available).
	ReasoningContent string
	// ToolCalls contains any tool calls the model wants to make.
	ToolCalls []*ToolCallInfo
	// FinishReason indicates why generation stopped.
	FinishReason FinishReason
	// Citations are external sources referenced in the response.
	Citations []string
	// Usage contains token usage information.
	Usage Usage
	// Model is the actual model that was used.
	Model string
	// Created is when the response was generated.
	Created time.Time
	// SystemFingerprint identifies the backend configuration.
	SystemFingerprint string
}

// HasToolCalls returns true if the response contains tool calls.
func (r *ChatResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// CompleteChat performs a blocking chat completion.
func (c *Client) CompleteChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	protoReq := req.Build(c.config.DefaultModel)

	resp, err := c.chat.GetCompletion(ctx, protoReq)
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return chatResponseFromProto(resp), nil
}

func chatResponseFromProto(resp *v1.GetChatCompletionResponse) *ChatResponse {
	result := &ChatResponse{
		ID:                resp.GetId(),
		Citations:         resp.GetCitations(),
		Usage:             usageFromProto(resp.GetUsage()),
		Model:             resp.GetModel(),
		SystemFingerprint: resp.GetSystemFingerprint(),
	}

	if resp.GetCreated() != nil {
		result.Created = resp.GetCreated().AsTime()
	}

	// Extract from first output (typically only one)
	if len(resp.GetOutputs()) > 0 {
		output := resp.GetOutputs()[0]
		result.FinishReason = finishReasonFromProto(output.GetFinishReason())

		if msg := output.GetMessage(); msg != nil {
			result.Content = msg.GetContent()
			result.ReasoningContent = msg.GetReasoningContent()

			for _, tc := range msg.GetToolCalls() {
				result.ToolCalls = append(result.ToolCalls, toolCallFromProto(tc))
			}
		}
	}

	return result
}

// ChatChunk represents a streaming chunk of a chat response.
type ChatChunk struct {
	// ID is the response ID.
	ID string
	// Delta is the incremental content.
	Delta string
	// ReasoningDelta is the incremental reasoning content.
	ReasoningDelta string
	// ToolCalls contains incremental tool call information.
	ToolCalls []*ToolCallInfo
	// FinishReason is set on the final chunk.
	FinishReason FinishReason
	// Citations are populated on the final chunk.
	Citations []string
	// Usage is updated on each chunk.
	Usage Usage
	// Model is the actual model used.
	Model string
}

// ChunkStream is an iterator for streaming chat chunks.
type ChunkStream struct {
	stream v1.Chat_GetCompletionChunkClient
	err    error
}

// Next returns the next chunk, or io.EOF when done.
// Any error other than io.EOF indicates a failure.
func (s *ChunkStream) Next() (*ChatChunk, error) {
	if s.err != nil {
		return nil, s.err
	}

	chunk, err := s.stream.Recv()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		s.err = FromGRPCError(err)
		return nil, s.err
	}

	return chunkFromProto(chunk), nil
}

// Close closes the stream.
func (s *ChunkStream) Close() error {
	// gRPC streams are closed automatically when the context is canceled
	// or when the server sends EOF. We just need to drain any remaining
	// messages to be safe.
	return nil
}

// Err returns any error that occurred during streaming.
func (s *ChunkStream) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func chunkFromProto(chunk *v1.GetChatCompletionChunk) *ChatChunk {
	result := &ChatChunk{
		ID:        chunk.GetId(),
		Citations: chunk.GetCitations(),
		Usage:     usageFromProto(chunk.GetUsage()),
		Model:     chunk.GetModel(),
	}

	// Extract from first output chunk
	if len(chunk.GetOutputs()) > 0 {
		output := chunk.GetOutputs()[0]
		result.FinishReason = finishReasonFromProto(output.GetFinishReason())

		if delta := output.GetDelta(); delta != nil {
			result.Delta = delta.GetContent()
			result.ReasoningDelta = delta.GetReasoningContent()

			for _, tc := range delta.GetToolCalls() {
				result.ToolCalls = append(result.ToolCalls, toolCallFromProto(tc))
			}
		}
	}

	return result
}

// StreamChat starts a streaming chat completion.
func (c *Client) StreamChat(ctx context.Context, req *ChatRequest) (*ChunkStream, error) {
	protoReq := req.Build(c.config.DefaultModel)

	stream, err := c.chat.GetCompletionChunk(ctx, protoReq)
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return &ChunkStream{stream: stream}, nil
}

// DeferredStatus represents the status of a deferred completion.
type DeferredStatus string

const (
	// DeferredStatusPending indicates the request is still processing.
	DeferredStatusPending DeferredStatus = "pending"
	// DeferredStatusCompleted indicates the request completed successfully.
	DeferredStatusCompleted DeferredStatus = "completed"
	// DeferredStatusFailed indicates the request failed.
	DeferredStatusFailed DeferredStatus = "failed"
)

// DeferredResponse contains the result of a deferred completion.
type DeferredResponse struct {
	// Status is the current status.
	Status DeferredStatus
	// Response is the chat response (only set when completed).
	Response *ChatResponse
	// Error is the error message (only set when failed).
	Error string
}

// StartDeferred starts a deferred (async) chat completion.
// Returns the request ID which can be used to poll for results.
func (c *Client) StartDeferred(ctx context.Context, req *ChatRequest) (string, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	protoReq := req.Build(c.config.DefaultModel)

	resp, err := c.chat.StartDeferredCompletion(ctx, protoReq)
	if err != nil {
		return "", FromGRPCError(err)
	}

	return resp.GetRequestId(), nil
}

// GetDeferred retrieves the status/result of a deferred completion.
func (c *Client) GetDeferred(ctx context.Context, requestID string) (*DeferredResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.chat.GetDeferredCompletion(ctx, &v1.GetDeferredRequest{
		RequestId: requestID,
	})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	result := &DeferredResponse{}

	switch resp.GetStatus() {
	case v1.DeferredStatus_PENDING:
		result.Status = DeferredStatusPending
	case v1.DeferredStatus_DONE:
		result.Status = DeferredStatusCompleted
		if resp.GetResponse() != nil {
			result.Response = chatResponseFromProto(resp.GetResponse())
		}
	case v1.DeferredStatus_EXPIRED:
		result.Status = DeferredStatusFailed
		result.Error = "deferred completion expired"
	}

	return result, nil
}

// WaitForDeferred polls for a deferred completion until it completes or times out.
func (c *Client) WaitForDeferred(ctx context.Context, requestID string, pollInterval, timeout time.Duration) (*ChatResponse, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := c.GetDeferred(ctx, requestID)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case DeferredStatusCompleted:
			return resp.Response, nil
		case DeferredStatusFailed:
			return nil, &Error{
				Code:    ErrServerError,
				Message: "deferred completion failed: " + resp.Error,
			}
		}

		// Still pending, wait before polling again
		select {
		case <-ctx.Done():
			return nil, FromGRPCError(ctx.Err())
		case <-time.After(pollInterval):
			// Continue polling
		}
	}

	return nil, &Error{
		Code:    ErrTimeout,
		Message: "timeout waiting for deferred completion",
	}
}

// GetStoredCompletion retrieves a stored completion by response ID.
func (c *Client) GetStoredCompletion(ctx context.Context, responseID string) (*ChatResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	resp, err := c.chat.GetStoredCompletion(ctx, &v1.GetStoredCompletionRequest{
		ResponseId: responseID,
	})
	if err != nil {
		return nil, FromGRPCError(err)
	}

	return chatResponseFromProto(resp), nil
}

// DeleteStoredCompletion deletes a stored completion by response ID.
func (c *Client) DeleteStoredCompletion(ctx context.Context, responseID string) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	_, err := c.chat.DeleteStoredCompletion(ctx, &v1.DeleteStoredCompletionRequest{
		ResponseId: responseID,
	})
	if err != nil {
		return FromGRPCError(err)
	}

	return nil
}
