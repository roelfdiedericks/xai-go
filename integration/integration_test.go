// Package integration contains integration tests that require a valid XAI_APIKEY.
package integration

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	xai "github.com/roelfdiedericks/xai-go"
)

// getClient creates a client from environment or skips the test.
func getClient(t *testing.T) *xai.Client {
	t.Helper()
	apiKey := os.Getenv("XAI_APIKEY")
	if apiKey == "" {
		t.Skip("XAI_APIKEY not set, skipping integration test")
	}
	client, err := xai.FromEnv()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

func TestGetAPIKeyInfo(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.GetAPIKeyInfo(ctx)
	if err != nil {
		t.Fatalf("GetAPIKeyInfo failed: %v", err)
	}

	if info.KeyID == "" {
		t.Error("KeyID should not be empty")
	}
	if info.RedactedKey == "" {
		t.Error("RedactedKey should not be empty")
	}
	t.Logf("API Key: %s (ID: %s)", info.RedactedKey, info.KeyID)
	t.Logf("Status: %s, ACLs: %v", info.Status, info.ACLs)
}

func TestListModels(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}

	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	for _, m := range models {
		t.Logf("Model: %s (aliases: %v)", m.Name, m.Aliases)
	}
}

func TestCompleteChat(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant."}).
		UserMessage(xai.UserContent{Text: "Say 'hello' and nothing else."}).
		WithMaxTokens(50)

	resp, err := client.CompleteChat(ctx, req)
	if err != nil {
		t.Fatalf("CompleteChat failed: %v", err)
	}

	if resp.ID == "" {
		t.Error("Response ID should not be empty")
	}
	if resp.Content == "" {
		t.Error("Response content should not be empty")
	}

	t.Logf("Response ID: %s", resp.ID)
	t.Logf("Content: %s", resp.Content)
	t.Logf("Finish reason: %s", resp.FinishReason)
	t.Logf("Usage: prompt=%d, completion=%d, total=%d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
}

func TestStreamChat(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant."}).
		UserMessage(xai.UserContent{Text: "Count from 1 to 5."}).
		WithMaxTokens(100)

	stream, err := client.StreamChat(ctx, req)
	if err != nil {
		t.Fatalf("StreamChat failed: %v", err)
	}

	var content string
	chunkCount := 0
	for {
		chunk, err := stream.Next()
		if err != nil {
			break
		}
		chunkCount++
		content += chunk.Delta
	}

	if stream.Err() != nil {
		t.Fatalf("Stream error: %v", stream.Err())
	}

	if content == "" {
		t.Error("Streamed content should not be empty")
	}
	if chunkCount == 0 {
		t.Error("Expected at least one chunk")
	}

	t.Logf("Received %d chunks", chunkCount)
	t.Logf("Streamed content: %s", content)
}

func TestTokenize(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Tokenize(ctx, client.DefaultModel(), "Hello, world!")
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	if resp.TokenCount() == 0 {
		t.Error("Expected at least one token")
	}

	t.Logf("Token count: %d", resp.TokenCount())
	for _, tok := range resp.Tokens {
		t.Logf("  Token %d: %q", tok.TokenID, tok.StringToken)
	}
}

func TestListEmbeddingModels(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := client.ListEmbeddingModels(ctx)
	if err != nil {
		t.Fatalf("ListEmbeddingModels failed: %v", err)
	}

	if len(models) == 0 {
		t.Skip("No embedding models available")
	}

	for _, m := range models {
		t.Logf("Embedding model: %s", m.Name)
	}
}

// TestChatMultiTurnWithHistory tests multi-turn conversation using full message history.
// This is the traditional approach where the client sends all previous messages with each request.
func TestChatMultiTurnWithHistory(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// First turn: establish context
	req1 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant. Be concise."}).
		UserMessage(xai.UserContent{Text: "My name is Alice. Remember this."}).
		WithMaxTokens(100)

	resp1, err := client.CompleteChat(ctx, req1)
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}
	t.Logf("Turn 1 response: %s", resp1.Content)

	// Second turn: include full history
	req2 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant. Be concise."}).
		UserMessage(xai.UserContent{Text: "My name is Alice. Remember this."}).
		AssistantMessage(xai.AssistantContent{Text: resp1.Content}).
		UserMessage(xai.UserContent{Text: "What is my name?"}).
		WithMaxTokens(50)

	resp2, err := client.CompleteChat(ctx, req2)
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}
	t.Logf("Turn 2 response: %s", resp2.Content)

	// Verify the model remembers the name
	if !containsIgnoreCase(resp2.Content, "Alice") {
		t.Errorf("Expected response to contain 'Alice', got: %s", resp2.Content)
	}
}

// TestChatMultiTurnWithResponseId tests multi-turn conversation using previous_response_id.
// This uses xAI's server-side context storage to maintain conversation state.
func TestChatMultiTurnWithResponseId(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// First turn: enable storage and establish context
	req1 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant. Be concise."}).
		UserMessage(xai.UserContent{Text: "My name is Bob. Remember this."}).
		WithStoreMessages(true).
		WithMaxTokens(100)

	resp1, err := client.CompleteChat(ctx, req1)
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}
	t.Logf("Turn 1 response ID: %s", resp1.ID)
	t.Logf("Turn 1 response: %s", resp1.Content)

	if resp1.ID == "" {
		t.Fatal("Expected response ID when store_messages=true, got empty string")
	}

	// Second turn: use previous_response_id, only send new message
	// Note: WithStoreMessages(true) is needed to continue the chain
	req2 := xai.NewChatRequest().
		WithPreviousResponseId(resp1.ID).
		WithStoreMessages(true).
		UserMessage(xai.UserContent{Text: "What is my name?"}).
		WithMaxTokens(50)

	resp2, err := client.CompleteChat(ctx, req2)
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}
	t.Logf("Turn 2 response ID: %s", resp2.ID)
	t.Logf("Turn 2 response: %s", resp2.Content)

	// Verify the model remembers the name from server-side context
	if !containsIgnoreCase(resp2.Content, "Bob") {
		t.Errorf("Expected response to contain 'Bob', got: %s", resp2.Content)
	}

	// Third turn: chain from second response
	req3 := xai.NewChatRequest().
		WithPreviousResponseId(resp2.ID).
		WithStoreMessages(true).
		UserMessage(xai.UserContent{Text: "Say my name backwards."}).
		WithMaxTokens(50)

	resp3, err := client.CompleteChat(ctx, req3)
	if err != nil {
		t.Fatalf("Third turn failed: %v", err)
	}
	t.Logf("Turn 3 response ID: %s", resp3.ID)
	t.Logf("Turn 3 response: %s", resp3.Content)

	// Verify continuity - should reference Bob backwards (boB)
	if !containsIgnoreCase(resp3.Content, "bob") && !containsIgnoreCase(resp3.Content, "boB") {
		t.Logf("Note: Response may not contain exact 'boB', got: %s", resp3.Content)
	}
}

// TestChatWithResponseIdStreaming tests streaming with previous_response_id.
func TestChatWithResponseIdStreaming(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// First turn: enable storage
	req1 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a helpful assistant. Be concise."}).
		UserMessage(xai.UserContent{Text: "The secret code is 42. Remember it."}).
		WithStoreMessages(true).
		WithMaxTokens(100)

	stream1, err := client.StreamChat(ctx, req1)
	if err != nil {
		t.Fatalf("First stream failed: %v", err)
	}

	var responseId string
	var content1 string
	for {
		chunk, err := stream1.Next()
		if err != nil {
			break
		}
		if chunk.ID != "" {
			responseId = chunk.ID
		}
		content1 += chunk.Delta
	}
	t.Logf("Turn 1 response ID: %s", responseId)
	t.Logf("Turn 1 content: %s", content1)

	if responseId == "" {
		t.Fatal("Expected response ID from streaming with store_messages=true")
	}

	// Second turn: use previous_response_id with streaming
	// Note: WithStoreMessages(true) would be needed to continue the chain further
	req2 := xai.NewChatRequest().
		WithPreviousResponseId(responseId).
		UserMessage(xai.UserContent{Text: "What is the secret code?"}).
		WithMaxTokens(50)

	stream2, err := client.StreamChat(ctx, req2)
	if err != nil {
		t.Fatalf("Second stream failed: %v", err)
	}

	var content2 string
	for {
		chunk, err := stream2.Next()
		if err != nil {
			break
		}
		content2 += chunk.Delta
	}
	t.Logf("Turn 2 content: %s", content2)

	// Verify the model remembers the code
	if !containsIgnoreCase(content2, "42") {
		t.Errorf("Expected response to contain '42', got: %s", content2)
	}
}

// TestChatWithToolCallHistory tests reconstructing conversation history that includes
// assistant tool calls and tool results. This validates the AssistantContent.ToolCalls
// and ToolContent types work correctly for history reconstruction.
func TestChatWithToolCallHistory(t *testing.T) {
	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Define a simple calculator tool
	addTool := xai.NewFunctionTool("add_numbers", "Add two numbers together").
		WithParameters(`{"type":"object","properties":{"a":{"type":"number","description":"First number"},"b":{"type":"number","description":"Second number"}},"required":["a","b"]}`)

	// First turn: ask model to use the tool
	req1 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a calculator assistant. Always use the add_numbers tool when asked to add."}).
		UserMessage(xai.UserContent{Text: "What is 2 + 3?"}).
		AddTool(addTool).
		WithToolChoice(xai.ToolChoiceRequired).
		WithMaxTokens(100)

	resp1, err := client.CompleteChat(ctx, req1)
	if err != nil {
		t.Fatalf("First turn failed: %v", err)
	}

	if !resp1.HasToolCalls() {
		t.Fatalf("Expected tool call in response, got none. Content: %s", resp1.Content)
	}

	tc := resp1.ToolCalls[0]
	if tc.Function == nil {
		t.Fatal("Tool call has no function")
	}
	t.Logf("Tool call ID: %s", tc.ID)
	t.Logf("Tool call: %s(%s)", tc.Function.Name, tc.Function.Arguments)

	// Second turn: reconstruct history with tool call + result, then ask follow-up
	req2 := xai.NewChatRequest().
		SystemMessage(xai.SystemContent{Text: "You are a calculator assistant. Always use the add_numbers tool when asked to add."}).
		UserMessage(xai.UserContent{Text: "What is 2 + 3?"}).
		AssistantMessage(xai.AssistantContent{
			Text: resp1.Content, // may be empty when tool is called
			ToolCalls: []xai.HistoryToolCall{{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}},
		}).
		ToolResult(xai.ToolContent{CallID: tc.ID, Result: "5"}).
		UserMessage(xai.UserContent{Text: "Great, what was that result again?"}).
		AddTool(addTool).
		WithMaxTokens(100)

	resp2, err := client.CompleteChat(ctx, req2)
	if err != nil {
		t.Fatalf("Second turn failed: %v", err)
	}
	t.Logf("Response: %s", resp2.Content)

	// Model should reference the result "5"
	if !containsIgnoreCase(resp2.Content, "5") {
		t.Errorf("Expected response to reference '5', got: %s", resp2.Content)
	}
}

// TestContextWindowBehavior documents how xAI handles oversized requests.
// FINDING: xAI does NOT return context window errors - they silently truncate using a sliding window.
// This test verifies this behavior and shows how to detect truncation via token counts.
// Run explicitly with: go test -v -run TestContextWindowBehavior ./integration/...
func TestContextWindowBehavior(t *testing.T) {
	t.Skip("Documentation test - run manually with: go test -v -run TestContextWindowBehavior ./integration/...")

	client := getClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create a large message (~150k tokens, exceeds 128k context of grok-4-1-fast)
	largeText := strings.Repeat("word ", 150000) // ~150k tokens

	req := xai.NewChatRequest().
		WithModel("grok-4-1-fast"). // 128k context window
		SystemMessage(xai.SystemContent{Text: "Count approximately how many times the word 'word' appears."}).
		UserMessage(xai.UserContent{Text: largeText}).
		WithMaxTokens(100)

	resp, err := client.CompleteChat(ctx, req)

	// DOCUMENTED FINDING: xAI does NOT return errors for oversized context
	// Instead, they silently truncate using a sliding window approach
	if err != nil {
		var xaiErr *xai.Error
		if errors.As(err, &xaiErr) {
			t.Logf("Got error (unexpected): Code=%s, Message=%s", xaiErr.Code.String(), xaiErr.Message)
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Logf("=== xAI CONTEXT WINDOW BEHAVIOR ===")
	t.Logf("Request succeeded (no error returned)")
	t.Logf("Prompt tokens used: %d", resp.Usage.PromptTokens)
	t.Logf("Expected tokens: ~150,000")
	t.Logf("Response: %s", resp.Content)

	// Check if truncation occurred by comparing token counts
	if resp.Usage.PromptTokens < 140000 {
		t.Logf("TRUNCATION DETECTED: Only %d prompt tokens used (expected ~150k)", resp.Usage.PromptTokens)
		t.Logf("xAI silently truncated the request to fit context window")
	}

	// Document the finding
	t.Logf("\n=== FINDING FOR GOCLAW ===")
	t.Logf("xAI does NOT return context window exceeded errors")
	t.Logf("Detection method: Compare resp.Usage.PromptTokens against expected")
	t.Logf("Or pre-flight check: Use client.Tokenize() before sending")
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			findIgnoreCase(s, substr))
}

func findIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
