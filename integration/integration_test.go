// Package integration contains integration tests that require a valid XAI_APIKEY.
package integration

import (
	"context"
	"os"
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
		SystemMessage("You are a helpful assistant.").
		UserMessage("Say 'hello' and nothing else.").
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
		SystemMessage("You are a helpful assistant.").
		UserMessage("Count from 1 to 5.").
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
